package liveapi

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"hsanalyzer/pkg/contracts/overlay"
)

type Server struct {
	mu       sync.RWMutex
	snapshot overlay.Snapshot
	clients  map[*wsClient]struct{}
}

type wsClient struct {
	conn net.Conn
	send chan []byte
}

func NewServer(initial overlay.Snapshot) *Server {
	initial.Normalize()
	return &Server{
		snapshot: initial,
		clients:  make(map[*wsClient]struct{}),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/live/overlay-snapshot", s.handleSnapshot)
	mux.HandleFunc("/ws/live", s.handleWebSocket)
	return mux
}

func (s *Server) Publish(snapshot overlay.Snapshot) error {
	if s == nil {
		return errors.New("live api server is nil")
	}

	snapshot.Normalize()
	if err := snapshot.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	previous := s.snapshot.Clone()
	s.snapshot = snapshot.Clone()
	clients := s.cloneClientsLocked()
	s.mu.Unlock()

	var messages [][]byte
	if previous.Match.MatchID != "" && snapshot.Match.MatchID != "" && previous.Match.MatchID != snapshot.Match.MatchID {
		messages = append(messages, mustMarshalEnvelope(overlay.Envelope{
			Type:          overlay.MessageTypeReset,
			SchemaVersion: overlay.SchemaVersionV1,
			MatchID:       previous.Match.MatchID,
		}))
	}

	messageType := overlay.MessageTypeDelta
	if previous.Match.MatchID == "" || previous.Match.MatchID != snapshot.Match.MatchID {
		messageType = overlay.MessageTypeSnapshot
	}
	messages = append(messages, mustMarshalEnvelope(overlay.Envelope{
		Type:          messageType,
		SchemaVersion: overlay.SchemaVersionV1,
		Delta:         &snapshot,
		Snapshot:      &snapshot,
	}))
	messages = append(messages, mustMarshalEnvelope(overlay.Envelope{
		Type:          overlay.MessageTypeTransport,
		SchemaVersion: overlay.SchemaVersionV1,
		Health:        &snapshot.Health,
	}))

	for _, payload := range messages {
		for _, client := range clients {
			client.enqueue(payload)
		}
	}
	return nil
}

func (s *Server) Snapshot() overlay.Snapshot {
	if s == nil {
		return overlay.NewSnapshot()
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot.Clone()
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, s.Snapshot())
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !isWebSocketUpgrade(r) {
		http.Error(w, "upgrade required", http.StatusUpgradeRequired)
		return
	}

	conn, err := upgradeWebSocket(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client := &wsClient{
		conn: conn,
		send: make(chan []byte, 16),
	}

	s.mu.Lock()
	s.clients[client] = struct{}{}
	s.mu.Unlock()

	go client.writeLoop(func() {
		s.mu.Lock()
		delete(s.clients, client)
		s.mu.Unlock()
	})

	snapshot := s.Snapshot()
	client.enqueue(mustMarshalEnvelope(overlay.Envelope{
		Type:          overlay.MessageTypeSnapshot,
		SchemaVersion: overlay.SchemaVersionV1,
		Snapshot:      &snapshot,
	}))
	client.enqueue(mustMarshalEnvelope(overlay.Envelope{
		Type:          overlay.MessageTypeTransport,
		SchemaVersion: overlay.SchemaVersionV1,
		Health:        &snapshot.Health,
	}))
}

func (s *Server) cloneClientsLocked() []*wsClient {
	result := make([]*wsClient, 0, len(s.clients))
	for client := range s.clients {
		result = append(result, client)
	}
	return result
}

func (c *wsClient) enqueue(payload []byte) {
	if c == nil {
		return
	}
	select {
	case c.send <- payload:
	default:
		// Если клиент не успевает читать, не блокируем мониторинг и закрываем соединение.
		close(c.send)
	}
}

func (c *wsClient) writeLoop(onClose func()) {
	defer func() {
		if onClose != nil {
			onClose()
		}
		_ = c.conn.Close()
	}()

	for payload := range c.send {
		if err := writeTextFrame(c.conn, payload); err != nil {
			return
		}
	}
}

func isWebSocketUpgrade(r *http.Request) bool {
	if r == nil {
		return false
	}
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func upgradeWebSocket(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("websocket hijacking is not supported")
	}

	key := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if key == "" {
		return nil, errors.New("missing Sec-WebSocket-Key header")
	}

	conn, buffer, err := hijacker.Hijack()
	if err != nil {
		return nil, fmt.Errorf("hijack websocket: %w", err)
	}

	acceptKey := buildWebSocketAcceptKey(key)
	response := bytes.NewBuffer(nil)
	fmt.Fprintf(response, "HTTP/1.1 101 Switching Protocols\r\n")
	fmt.Fprintf(response, "Upgrade: websocket\r\n")
	fmt.Fprintf(response, "Connection: Upgrade\r\n")
	fmt.Fprintf(response, "Sec-WebSocket-Accept: %s\r\n", acceptKey)
	fmt.Fprintf(response, "\r\n")

	if _, err := buffer.Write(response.Bytes()); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := buffer.Flush(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func buildWebSocketAcceptKey(key string) string {
	h := sha1.New() // #nosec G401 - websocket handshake requires SHA-1 by protocol
	_, _ = h.Write([]byte(key))
	_, _ = h.Write([]byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func mustMarshalEnvelope(envelope overlay.Envelope) []byte {
	if envelope.Type == overlay.MessageTypeSnapshot && envelope.Snapshot != nil {
		envelope.Delta = nil
	}
	if envelope.Type == overlay.MessageTypeDelta && envelope.Delta != nil {
		envelope.Snapshot = nil
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		panic(err)
	}
	return payload
}

func writeTextFrame(conn net.Conn, payload []byte) error {
	if conn == nil {
		return errors.New("websocket connection is nil")
	}

	header := []byte{0x81}
	length := len(payload)
	switch {
	case length <= 125:
		header = append(header, byte(length))
	case length <= 65535:
		header = append(header, 126, byte(length>>8), byte(length))
	default:
		header = append(header,
			127,
			byte(length>>56),
			byte(length>>48),
			byte(length>>40),
			byte(length>>32),
			byte(length>>24),
			byte(length>>16),
			byte(length>>8),
			byte(length),
		)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return err
	}
	if _, err := conn.Write(header); err != nil {
		return err
	}
	_, err := conn.Write(payload)
	return err
}

func DialForTest(addr string, timeout time.Duration) (net.Conn, *bufio.Reader, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, nil, err
	}
	reader := bufio.NewReader(conn)
	return conn, reader, nil
}

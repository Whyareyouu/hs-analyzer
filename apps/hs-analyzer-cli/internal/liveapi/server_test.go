package liveapi

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hsanalyzer/pkg/contracts/overlay"
)

func TestSnapshotRouteReturnsOverlaySnapshot(t *testing.T) {
	t.Parallel()

	snapshot := overlay.NewSnapshot()
	snapshot.Match = overlay.Match{MatchID: "game:1", GameID: "1", Status: overlay.MatchStatusActive}

	server := NewServer(snapshot)
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/overlay-snapshot", nil)
	server.Routes().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", resp.Code, resp.Body.String())
	}

	var decoded overlay.Snapshot
	if err := json.Unmarshal(resp.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if decoded.Match.MatchID != "game:1" {
		t.Fatalf("unexpected snapshot: %#v", decoded)
	}
}

func TestWebSocketPushesInitialSnapshotAndDelta(t *testing.T) {
	t.Parallel()

	initial := overlay.NewSnapshot()
	initial.Match = overlay.Match{MatchID: "game:1", GameID: "1", Status: overlay.MatchStatusActive}

	server := NewServer(initial)
	httpServer := httptest.NewServer(server.Routes())
	defer httpServer.Close()

	conn, reader := connectWebSocket(t, httpServer.Listener.Addr().String(), "/ws/live")
	defer conn.Close()

	first := readEnvelope(t, conn, reader)
	if first.Type != overlay.MessageTypeSnapshot {
		t.Fatalf("expected initial snapshot message, got %#v", first)
	}

	health := readEnvelope(t, conn, reader)
	if health.Type != overlay.MessageTypeTransport {
		t.Fatalf("expected transport health message, got %#v", health)
	}

	next := initial.Clone()
	next.OpponentPlayedCards = []overlay.OpponentPlayedCard{
		{CardID: "CARD_1", Name: "Карта", ManaCost: 2, PlayCount: 1},
	}
	if err := server.Publish(next); err != nil {
		t.Fatalf("publish snapshot: %v", err)
	}

	update := readEnvelope(t, conn, reader)
	if update.Type != overlay.MessageTypeDelta && update.Type != overlay.MessageTypeSnapshot {
		t.Fatalf("expected delta/snapshot update, got %#v", update)
	}
	payload := update.Delta
	if payload == nil {
		payload = update.Snapshot
	}
	if payload == nil || len(payload.OpponentPlayedCards) != 1 {
		t.Fatalf("unexpected update payload: %#v", update)
	}
}

func TestWebSocketSendsResetWhenMatchChanges(t *testing.T) {
	t.Parallel()

	initial := overlay.NewSnapshot()
	initial.Match = overlay.Match{MatchID: "game:1", GameID: "1", Status: overlay.MatchStatusActive}

	server := NewServer(initial)
	httpServer := httptest.NewServer(server.Routes())
	defer httpServer.Close()

	conn, reader := connectWebSocket(t, httpServer.Listener.Addr().String(), "/ws/live")
	defer conn.Close()

	_ = readEnvelope(t, conn, reader)
	_ = readEnvelope(t, conn, reader)

	next := initial.Clone()
	next.Match.MatchID = "game:2"
	next.Match.GameID = "2"
	if err := server.Publish(next); err != nil {
		t.Fatalf("publish snapshot: %v", err)
	}

	reset := readEnvelope(t, conn, reader)
	if reset.Type != overlay.MessageTypeReset {
		t.Fatalf("expected reset message, got %#v", reset)
	}
	if reset.MatchID != "game:1" {
		t.Fatalf("unexpected reset payload: %#v", reset)
	}
}

func connectWebSocket(t *testing.T, address, path string) (net.Conn, *bufio.Reader) {
	t.Helper()

	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}

	key := base64.StdEncoding.EncodeToString([]byte("test-websocket-key"))
	request := fmt.Sprintf(
		"GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n",
		path,
		address,
		key,
	)
	if _, err := conn.Write([]byte(request)); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read handshake status: %v", err)
	}
	if !strings.Contains(statusLine, "101") {
		t.Fatalf("unexpected handshake status: %s", statusLine)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read handshake headers: %v", err)
		}
		if line == "\r\n" {
			break
		}
	}
	return conn, reader
}

func readEnvelope(t *testing.T, conn net.Conn, reader *bufio.Reader) overlay.Envelope {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	payload := readTextFrame(t, reader)
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		t.Fatalf("reset read deadline: %v", err)
	}

	var envelope overlay.Envelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("decode websocket message: %v payload=%s", err, string(payload))
	}
	if err := envelope.Validate(); err != nil {
		t.Fatalf("validate websocket message: %v payload=%s", err, string(payload))
	}
	return envelope
}

func readTextFrame(t *testing.T, reader *bufio.Reader) []byte {
	t.Helper()

	first, err := reader.ReadByte()
	if err != nil {
		t.Fatalf("read first frame byte: %v", err)
	}
	if first != 0x81 {
		t.Fatalf("unexpected first frame byte: %#x", first)
	}

	second, err := reader.ReadByte()
	if err != nil {
		t.Fatalf("read second frame byte: %v", err)
	}

	length := int(second & 0x7F)
	switch length {
	case 126:
		high, err := reader.ReadByte()
		if err != nil {
			t.Fatalf("read extended frame high byte: %v", err)
		}
		low, err := reader.ReadByte()
		if err != nil {
			t.Fatalf("read extended frame low byte: %v", err)
		}
		length = int(high)<<8 | int(low)
	case 127:
		var lengthBytes [8]byte
		if _, err := io.ReadFull(reader, lengthBytes[:]); err != nil {
			t.Fatalf("read 64-bit frame length: %v", err)
		}
		length = 0
		for _, b := range lengthBytes {
			length = (length << 8) | int(b)
		}
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		t.Fatalf("read websocket payload: %v", err)
	}
	return payload
}

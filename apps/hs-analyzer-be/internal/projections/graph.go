package projections

import (
	"fmt"
	"sort"

	"hsanalyzer/apps/hs-analyzer-be/internal/api"
)

type Node struct {
	ID   string
	Kind string
}

type Edge struct {
	From string
	To   string
	Kind string
}

type Graph struct {
	Nodes []Node
	Edges []Edge
}

func BuildGraph(matches []api.StoredMatch) Graph {
	nodes := make([]Node, 0)
	edges := make([]Edge, 0)

	for _, match := range matches {
		matchID := match.Payload.MatchID
		nodes = append(nodes, Node{ID: matchID, Kind: "match"})
		for _, event := range match.Payload.Events {
			eventID := fmt.Sprintf("%s:event:%d", matchID, event.Index)
			nodes = append(nodes, Node{ID: eventID, Kind: "event"})
			edges = append(edges, Edge{From: matchID, To: eventID, Kind: "HAS_EVENT"})
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].ID == nodes[j].ID {
			return nodes[i].Kind < nodes[j].Kind
		}
		return nodes[i].ID < nodes[j].ID
	})
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			if edges[i].To == edges[j].To {
				return edges[i].Kind < edges[j].Kind
			}
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})

	return Graph{Nodes: nodes, Edges: edges}
}

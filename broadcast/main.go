package main

import (
	"context"
	"encoding/json"
	"log"
	"maelstrom-broadcast/requests"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	mu := sync.RWMutex{}
	values := make([]int, 0)
	nodes := make(map[string]map[int]struct{})
	neighbors := make([]string, 0)
	counter := 0

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var req requests.Broadcast
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return &maelstrom.RPCError{Code: maelstrom.MalformedRequest, Text: err.Error()}
		}
		resBody := make(map[string]any)
		gossipReq := make(map[string]any)

		mu.Lock()
		values = append(values, req.Message)
		counter += 1
		gossipCounter := counter
		mu.Unlock()

		resBody["type"] = "broadcast_ok"
		gossipReq["type"] = "gossip"
		gossipReq["sender"] = n.ID()
		gossipReq["message"] = req.Message
		gossipReq["counter"] = gossipCounter
		for _, neighbor := range neighbors {
			go func(dest string) {
				for {
					err := n.Send(dest, gossipReq)
					if err == nil {
						break
					}
				}
			}(neighbor)
		}
		return n.Reply(msg, resBody)
	})

	n.Handle("gossip", func(msg maelstrom.Message) error {
		var req requests.Gossip
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return &maelstrom.RPCError{Code: maelstrom.MalformedRequest, Text: err.Error()}
		}

		mu.Lock()
		prev := nodes[req.Sender]
		if prev == nil {
			nodes[req.Sender] = make(map[int]struct{})
		}
		gossip := false
		if _, ok := nodes[req.Sender][req.Counter]; !ok {
			values = append(values, req.Message)
			nodes[req.Sender][req.Counter] = struct{}{}
			gossip = true
		}
		mu.Unlock()
		res := make(map[string]any)
		res["type"] = "gossip_ok"
		if gossip {
			gossipMsg := make(map[string]any)
			gossipMsg["type"] = "gossip"
			gossipMsg["counter"] = req.Counter
			gossipMsg["message"] = req.Message
			gossipMsg["sender"] = req.Sender
			for _, neighbor := range neighbors {
				go func(dest string) {
					for {
						ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
						defer cancel()
						msg, err := n.SyncRPC(ctx, dest, gossipMsg)
						if err == nil && msg.Type() == "gossip_ok" {
							break
						}
					}
				}(neighbor)
			}
		}
		return n.Reply(msg, res)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return &maelstrom.RPCError{Code: maelstrom.MalformedRequest, Text: err.Error()}
		}

		body["type"] = "read_ok"
		mu.RLock()
		body["messages"] = values
		mu.RUnlock()
		return n.Reply(msg, body)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var req requests.Topology
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return &maelstrom.RPCError{Code: maelstrom.MalformedRequest, Text: err.Error()}
		}

		mu.Lock()
		neighbors = req.Topology[n.ID()]
		mu.Unlock()
		res := make(map[string]any)
		res["type"] = "topology_ok"
		return n.Reply(msg, res)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

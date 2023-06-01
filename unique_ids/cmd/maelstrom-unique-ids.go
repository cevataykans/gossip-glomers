package main

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	nodeId := getNodeId()
	generator := SnowflakeGenerator(nodeId)

	node := maelstrom.NewNode()
	node.Handle("generate", func(msg maelstrom.Message) error {
		id := generator()
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "generate_ok"
		body["id"] = id
		return node.Reply(msg, body)
	})

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}

func getNodeId() uint64 {
	res, err := http.Get("http://localhost:13497/")
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}
	defer res.Body.Close()

	var nodeId uint64
	if err := binary.Read(res.Body, binary.LittleEndian, &nodeId); err != nil {
		log.Fatalf("err: %v\n", err)
	}
	log.Printf("Node id: %v\n", nodeId)
	return nodeId
}

func SnowflakeGenerator(nodeId uint64) func() uint64 {

	mu := sync.Mutex{}
	epoch := time.Now().Round(time.Millisecond)
	last_timestamp := time.Since(epoch).Milliseconds()
	counter := 0
	return func() uint64 {
		mu.Lock()
		defer mu.Unlock()

		diff := time.Since(epoch).Milliseconds()
		if diff == last_timestamp {
			counter = (counter + 1) & 1024
			if counter == 0 {
				for diff <= last_timestamp {
					diff = time.Since(epoch).Milliseconds()
				}
			}
		} else {
			counter = 0
		}
		last_timestamp = diff
		return (uint64(last_timestamp<<12) | uint64(nodeId<<10) | uint64(counter))
	}
}

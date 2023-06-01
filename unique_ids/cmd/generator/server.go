package main

import (
	"encoding/binary"
	"log"
	"net/http"
	"sync"
)

func main() {
	println("Node id server")

	curId := 0
	lock := sync.Mutex{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var nodeId uint64 = uint64(curId)
		curId = (curId + 1) % 4 // only support 4 nodes
		w.WriteHeader(http.StatusOK)
		if err := binary.Write(w, binary.LittleEndian, nodeId); err != nil {
			log.Fatalf("Cannot send node id: %v over http response, err: %v\n", nodeId, err)
		}
	})

	server := http.Server{
		Addr:    "localhost:13497",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

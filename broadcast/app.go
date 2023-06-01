package main

import "sync"

type gossipNode struct {
	counter   uint
	values    []int
	mu        sync.RWMutex
	neighbors map[string]Neighbor
}

func NewGossipNode() {

}

func Receive() {

}

func OnGossipRes() {

}

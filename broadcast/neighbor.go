package main

import "sync"

type Neighbor struct {
	dest string
	mu   sync.RWMutex
	msgs map[int]int
}

func SendGossip() {

}

func Acknowledge() {

}

func gossip_routine() {

}

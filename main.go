package main

import (
	"./cache"
	"./http"
	"./tcp"
)

func main() {
	c := cache.New("inmemory")
	go tcp.New(c).Listen()
	http.New(c).Listen()
}

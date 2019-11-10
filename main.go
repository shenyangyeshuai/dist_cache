package main

import (
	"./cache"
	"./cluster"
	"./http"
	"./tcp"
	"flag"
	"log"
)

func main() {
	node := flag.String("node", "127.0.0.1", "node address")
	clus := flag.String("cluster", "", "cluster address")
	flag.Parse()

	log.Println("node is", *node)
	log.Println("cluster is", *clus)

	c := cache.New("inmemory")
	n, e := cluster.New(*node, *clus)
	if e != nil {
		panic(e)
	}

	go tcp.New(c, n).Listen()
	http.New(c, n).Listen()
}

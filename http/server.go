package http

import (
	"../cache"
	"../cluster"
	"net/http"
)

type Server struct {
	cache cache.Cache
	node  cluster.Node
}

func New(c cache.Cache, n cluster.Node) *Server {
	return &Server{cache: c, node: n}
}

func (s *Server) Listen() {
	http.Handle("/cache/", s.cacheHandler())
	http.Handle("/status", s.statusHandler())
	http.Handle("/cluster", s.clusterHandler())
	http.Handle("/rebalance", s.rebalanceHandler())
	http.ListenAndServe(s.node.Addr()+":12345", nil)
}

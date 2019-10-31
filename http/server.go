package http

import (
	"../cache"
	"net/http"
)

type Server struct {
	cache cache.Cache
}

func New(c cache.Cache) *Server {
	return &Server{cache: c}
}

func (s *Server) Listen() {
	http.Handle("/cache/", s.cacheHandler())
	http.Handle("/status", s.statusHandler())
	http.ListenAndServe(":12345", nil)
}

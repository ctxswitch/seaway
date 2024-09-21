package handlers

import "net/http"

type Ping struct{}

func NewPingHandler() *Ping {
	return &Ping{}
}

func (h *Ping) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}

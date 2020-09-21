package rafthttp

import "net/http"

// Transporter -网络接口核心
type Transporter interface {
	Handler() http.Handler
}

type Transport struct {
}

func (t *Transport) Handler() http.Handler {
	mux := http.NewServeMux()

	return mux
}

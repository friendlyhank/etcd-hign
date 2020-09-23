package rafthttp

import (
	"fmt"
	"net/http"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"
)

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

func (t *Transport) Send(msgs []raftpb.Message) {
	for _, m := range msgs {
		fmt.Println(m)
	}
}

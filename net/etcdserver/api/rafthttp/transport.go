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
	streamHandler := newStreamHandler()
	mux := http.NewServeMux()
	mux.Handle(RaftStreamPrefix+"/", streamHandler)
	return mux
}

func (t *Transport) Send(msgs []raftpb.Message) {
	for _, m := range msgs {
		fmt.Println(m)
	}
}

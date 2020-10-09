package rafthttp

import (
	"fmt"
	"net/http"
	"path"
)

var (
	RaftPrefix       = "/raft"
	RaftStreamPrefix = path.Join(RaftPrefix, "stream")
)

type streamHandler struct {
}

func newStreamHandler() http.Handler {
	h := &streamHandler{}
	return h
}

func (h *streamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn := &outgoingConn{}
	fmt.Println(conn)
}

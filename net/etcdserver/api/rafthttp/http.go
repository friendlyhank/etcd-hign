package rafthttp

import (
	"fmt"
	"net/http"
	"path"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
)

var (
	RaftPrefix       = "/raft"
	RaftStreamPrefix = path.Join(RaftPrefix, "stream")
)

type peerGetter interface {
	Get(id types.ID) Peer
}

type streamHandler struct {
	lg         *zap.Logger
	tr         *Transport
	peerGetter peerGetter
	id         types.ID
}

func newStreamHandler(t *Transport, pg peerGetter, id types.ID) http.Handler {
	h := &streamHandler{
		lg:         t.Logger,
		tr:         t,
		peerGetter: pg,
		id:         id,
	}
	if h.lg == nil {
		h.lg = zap.NewNop()
	}
	return h
}

func (h *streamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fromStr := path.Base(r.URL.Path)
	from, err := types.IDFromString(fromStr)
	if err != nil {
		http.Error(w, "invalid from", http.StatusNotFound)
		return
	}
	//获取指定的peer
	p := h.peerGetter.Get(from)
	if p == nil {
		http.Error(w, "error sender not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	conn := &outgoingConn{
		Flusher: w.(http.Flusher),
		localID: h.tr.ID,
		peerID:  from,
	}
	fmt.Println(conn)
}

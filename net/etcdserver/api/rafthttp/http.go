package rafthttp

import (
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

	var t streamType
	switch path.Dir(r.URL.Path) {
	case streamTypeMsgAppV2.endpoint(h.lg):
		t = streamTypeMsgAppV2
	case streamTypeMessage.endpoint(h.lg):
		t = streamTypeMessage
	default:
		http.Error(w, "invalid path", http.StatusNotFound)
		return
	}

	fromStr := path.Base(r.URL.Path)
	from, err := types.IDFromString(fromStr)
	if err != nil {
		http.Error(w, "invalid from", http.StatusNotFound)
		return
	}
	//获取指定的peer TODO HANK写死编号 1849879258734672239
	p := h.peerGetter.Get(types.ID(1849879258734672239))
	if p == nil {
		http.Error(w, "error sender not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	conn := &outgoingConn{
		t:       t,
		Flusher: w.(http.Flusher),
		localID: h.tr.ID,
		peerID:  from,
	}
	p.attachOutgoingConn(conn)
}

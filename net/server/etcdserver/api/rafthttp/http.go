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

type pipelineHandler struct {
	lg      *zap.Logger
	localID types.ID
	tr      *Transport
	cid     types.ID
}

// newPipelineHandler returns a handler for handling raft messages
// from pipeline for RaftPrefix.
//
// The handler reads out the raft message from request body,
// and forwards it to the given raft state machine for processing.
func newPipelineHandler(t *Transport, cid types.ID) http.Handler {
	return &pipelineHandler{
		lg:      t.Logger,
		localID: t.ID,
		tr:      t,
		cid:     cid,
	}
}

func (h *pipelineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

type streamHandler struct {
	lg         *zap.Logger
	tr         *Transport
	peerGetter peerGetter
	id         types.ID
	cid        types.ID
}

func newStreamHandler(t *Transport, pg peerGetter, id, cid types.ID) http.Handler {
	h := &streamHandler{
		lg:         t.Logger,
		tr:         t,
		peerGetter: pg,
		id:         id,
		cid:        cid,
	}
	return h
}

func (h *streamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
	var t streamType
	switch path.Dir(r.URL.Path) {
	case streamTypeMsgAppV2.endpoint(h.lg):
		t = streamTypeMsgAppV2
	case streamTypeMessage.endpoint(h.lg):
		t = streamTypeMessage
	default:
		if h.lg != nil {
			h.lg.Debug(
				"ignored unexpected streaming request path",
				zap.String("local-member-id", h.tr.ID.String()),
				zap.String("remote-peer-id-stream-handler", h.id.String()),
				zap.String("path", r.URL.Path),
			)
		}
		http.Error(w, "invalid path", http.StatusNotFound)
		return
	}

	fromStr := path.Base(r.URL.Path)
	from, err := types.IDFromString(fromStr)
	if err != nil {
		if h.lg != nil {
			h.lg.Warn(
				"failed to parse path into ID",
				zap.String("local-member-id", h.tr.ID.String()),
				zap.String("remote-peer-id-stream-handler", h.id.String()),
				zap.String("path", fromStr),
				zap.Error(err),
			)
		}
		http.Error(w, "invalid from", http.StatusNotFound)
		return
	}
	//获取指定的peer
	p := h.peerGetter.Get(from)
	if p == nil {
		h.lg.Warn(
			"failed to find remote peer in cluster",
			zap.String("local-member-id", h.tr.ID.String()),
			zap.String("remote-peer-id-stream-handler", h.id.String()),
			zap.String("remote-peer-id-from", from.String()),
			zap.String("cluster-id", h.cid.String()),
		)
		http.Error(w, "error sender not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	conn := &outgoingConn{
		t:       t,
		Writer:  w,
		Flusher: w.(http.Flusher),
		localID: h.tr.ID,
		peerID:  from,
	}
	p.attachOutgoingConn(conn)
}

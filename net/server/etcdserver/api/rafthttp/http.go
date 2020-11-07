package rafthttp

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"

	pioutil "github.com/friendlyhank/etcd-hign/net/pkg/ioutil"
	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"go.uber.org/zap"
)

const (
	// connReadLimitByte limits the number of bytes
	// a single read can read out.
	//
	// 64KB should be large enough for not causing
	// throughput bottleneck as well as small enough
	// for not causing a read timeout.
	connReadLimitByte = 64 * 1024

	// snapshotLimitByte limits the snapshot size to 1TB
	snapshotLimitByte = 1 * 1024 * 1024 * 1024 * 1024
)

var (
	RaftPrefix       = "/raft"
	RaftStreamPrefix = path.Join(RaftPrefix, "stream")

	errIncompatibleVersion = errors.New("incompatible version")
	errClusterIDMismatch   = errors.New("cluster ID mismatch")
)

type peerGetter interface {
	Get(id types.ID) Peer
}

type pipelineHandler struct {
	lg      *zap.Logger
	localID types.ID
	tr      *Transport
	r       Raft
	cid     types.ID
}

// newPipelineHandler returns a handler for handling raft messages
// from pipeline for RaftPrefix.
//
// The handler reads out the raft message from request body,
// and forwards it to the given raft state machine for processing.
func newPipelineHandler(t *Transport, r Raft, cid types.ID) http.Handler {
	h := &pipelineHandler{
		lg:      t.Logger,
		localID: t.ID,
		tr:      t,
		r:       r,
		cid:     cid,
	}
	if h.lg == nil {
		h.lg = zap.NewNop()
	}
	return h
}

func (h *pipelineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("X-Etcd-Cluster-ID", h.cid.String())

	if err := checkClusterCompatibilityFromHeader(h.lg, h.localID, r.Header, h.cid); err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		return
	}

	// Limit the data size that could be read from the request body, which ensures that read from
	// connection will not time out accidentally due to possible blocking in underlying implementation.
	limitedr := pioutil.NewLimitedBufferReader(r.Body, connReadLimitByte)
	b, err := ioutil.ReadAll(limitedr)
	if err != nil {
		h.lg.Warn(
			"failed to read Raft message",
			zap.String("local-member-id", h.localID.String()),
			zap.Error(err),
		)
		http.Error(w, "error reading raft message", http.StatusBadRequest)
		return
	}

	var m raftpb.Message
	if err := m.Unmarshal(b); err != nil {
		h.lg.Warn(
			"failed to unmarshal Raft message",
			zap.String("local-member-id", h.localID.String()),
			zap.Error(err),
		)
		http.Error(w, "error unmarshalling raft message", http.StatusBadRequest)
		return
	}

	if err := h.r.Process(context.TODO(), m); err != nil {

	}
}

type streamHandler struct {
	lg         *zap.Logger
	tr         *Transport
	peerGetter peerGetter
	r          Raft
	id         types.ID
	cid        types.ID
}

func newStreamHandler(t *Transport, pg peerGetter, r Raft, id, cid types.ID) http.Handler {
	h := &streamHandler{
		lg:         t.Logger,
		tr:         t,
		peerGetter: pg,
		r:          r,
		id:         id,
		cid:        cid,
	}
	if h.lg == nil {
		h.lg = zap.NewNop()
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

	c := newCloseNotifier()
	conn := &outgoingConn{
		t:       t,
		Writer:  w,
		Flusher: w.(http.Flusher),
		Closer:  c,
		localID: h.tr.ID,
		peerID:  from,
	}
	p.attachOutgoingConn(conn)
	<-c.closeNotify()
}

// checkClusterCompatibilityFromHeader checks the cluster compatibility of
// the local member from the given header.
// It checks whether the version of local member is compatible with
// the versions in the header, and whether the cluster ID of local member
// matches the one in the header.
func checkClusterCompatibilityFromHeader(lg *zap.Logger, localID types.ID, header http.Header, cid types.ID) error {
	remoteName := header.Get("X-Server-From")

	remoteServer := serverVersion(header)
	remoteVs := ""
	if remoteServer != nil {
		remoteVs = remoteServer.String()
	}

	remoteMinClusterVer := minClusterVersion(header)
	remoteMinClusterVs := ""
	if remoteMinClusterVer != nil {
		remoteMinClusterVs = remoteMinClusterVer.String()
	}

	localServer, localMinCluster, err := checkVersionCompatibility(remoteName, remoteServer, remoteMinClusterVer)

	localVs := ""
	if localServer != nil {
		localVs = localServer.String()
	}
	localMinClusterVs := ""
	if localMinCluster != nil {
		localMinClusterVs = localMinCluster.String()
	}

	if err != nil {
		lg.Warn(
			"failed to check version compatibility",
			zap.String("local-member-id", localID.String()),
			zap.String("local-member-cluster-id", cid.String()),
			zap.String("local-member-server-version", localVs),
			zap.String("local-member-server-minimum-cluster-version", localMinClusterVs),
			zap.String("remote-peer-server-name", remoteName),
			zap.String("remote-peer-server-version", remoteVs),
			zap.String("remote-peer-server-minimum-cluster-version", remoteMinClusterVs),
			zap.Error(err),
		)
		return errIncompatibleVersion
	}

	if gcid := header.Get("X-Etcd-Cluster-ID"); gcid != cid.String() {
		lg.Warn(
			"request cluster ID mismatch",
			zap.String("local-member-id", localID.String()),
			zap.String("local-member-cluster-id", cid.String()),
			zap.String("local-member-server-version", localVs),
			zap.String("local-member-server-minimum-cluster-version", localMinClusterVs),
			zap.String("remote-peer-server-name", remoteName),
			zap.String("remote-peer-server-version", remoteVs),
			zap.String("remote-peer-server-minimum-cluster-version", remoteMinClusterVs),
			zap.String("remote-peer-cluster-id", gcid),
		)
		return errClusterIDMismatch
	}
	return nil
}

type closeNotifier struct {
	done chan struct{}
}

func newCloseNotifier() *closeNotifier {
	return &closeNotifier{
		done: make(chan struct{}),
	}
}

func (n *closeNotifier) Close() error {
	close(n.done)
	return nil
}

func (n *closeNotifier) closeNotify() <-chan struct{} { return n.done }

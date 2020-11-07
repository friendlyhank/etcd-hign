package rafthttp

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/friendlyhank/etcd-hign/net/server/etcdserver/api/snap"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/net/pkg/pbutil"
	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"
	"github.com/friendlyhank/etcd-hign/net/server/etcdserver/api/version"
)

//测试pipelineHandler
func TestServeRaftPrefix(t *testing.T) {
	testCases := []struct {
		method    string
		body      io.Reader
		p         Raft
		clusterID string

		wcode int
	}{
		{
			// bad method
			"GET",
			bytes.NewReader(pbutil.MustMarshal(&raftpb.Message{})),
			&fakeRaft{},
			"0",
			http.StatusMethodNotAllowed,
		},
		{
			// bad method
			"PUT",
			bytes.NewReader(pbutil.MustMarshal(&raftpb.Message{})),
			&fakeRaft{},
			"0",
			http.StatusMethodNotAllowed,
		},
		{
			// bad method
			"DELETE",
			bytes.NewReader(pbutil.MustMarshal(&raftpb.Message{})),
			&fakeRaft{},
			"0",
			http.StatusMethodNotAllowed,
		},
		{
			//bad request body
			"POST",
			&errReader{}, //io读取错误
			&fakeRaft{},
			"0",
			http.StatusBadRequest,
		},
		{
			//bad request protobuf
			"POST",
			strings.NewReader("malformed garbage"),
			&fakeRaft{},
			"0",
			http.StatusBadRequest,
		},
		{
			// good request, wrong cluster ID
			"POST",
			bytes.NewReader(
				pbutil.MustMarshal(&raftpb.Message{}),
			),
			&fakeRaft{},
			"1", //集群有唯一ID,集群中的每个节点有唯一ID
			http.StatusPreconditionFailed,
		},
		//{
		//	// good request, Processor failure
		//	"POST",
		//	bytes.NewReader(
		//		pbutil.MustMarshal(&raftpb.Message{}),
		//	),
		//	&fakeRaft{
		//		err: &resWriterToError{code: http.StatusForbidden}, //TODO 迟点再回来看这个干啥的
		//	},
		//	"0",
		//	http.StatusForbidden,
		//},
		//{
		//	// good request, Processor failure
		//	"POST",
		//	bytes.NewReader(
		//		pbutil.MustMarshal(&raftpb.Message{}),
		//	),
		//	&fakeRaft{
		//		err: &resWriterToError{code: http.StatusInternalServerError},
		//	},
		//	"0",
		//	http.StatusInternalServerError,
		//},
		//{
		//	// good request, Processor failure
		//	"POST",
		//	bytes.NewReader(
		//		pbutil.MustMarshal(&raftpb.Message{}),
		//	),
		//	&fakeRaft{err: errors.New("blah")},
		//	"0",
		//	http.StatusInternalServerError,
		//},
		{
			// good request
			"POST",
			bytes.NewReader(
				pbutil.MustMarshal(&raftpb.Message{}),
			),
			&fakeRaft{},
			"0",
			http.StatusNoContent,
		},
	}

	for i, tt := range testCases {
		req, err := http.NewRequest(tt.method, "foo", tt.body)
		if err != nil {
			t.Fatalf("#%d: could not create request: %#v", i, err)
		}
		req.Header.Set("X-Etcd-Cluster-ID", tt.clusterID)
		req.Header.Set("X-Server-Version", version.Version)
		rw := httptest.NewRecorder()
		h := newPipelineHandler(&Transport{Logger: zap.NewExample()}, tt.p, types.ID(0))

		// goroutine because the handler panics to disconnect on raft error
		donec := make(chan struct{})
		go func() {
			defer func() {
				recover()
				close(donec)
			}()
			h.ServeHTTP(rw, req) //可以这样写
		}()
		<-donec
		if rw.Code != tt.wcode {
			t.Errorf("#%d: got code=%d, want %d", i, rw.Code, tt.wcode)
		}
	}
}

func TestServeRaftStreamPrefix(t *testing.T) {
	tests := []struct {
		path  string
		wtype streamType
	}{
		{
			RaftStreamPrefix + "/message/1",
			streamTypeMessage,
		},
		{
			RaftStreamPrefix + "/msgapp/1",
			streamTypeMsgAppV2,
		},
	}

	for i, tt := range tests {
		req, err := http.NewRequest("GET", "http://localhost:2380"+tt.path, nil)
		if err != nil {
			t.Fatalf("#%d: could not create request: %#v", i, err)
		}
		req.Header.Set("X-Etcd-Cluster-ID", "1")
		req.Header.Set("X-Server-Version", version.Version)
		req.Header.Set("X-Raft-To", "2")

		peer := newFakePeer()
		peerGetter := &fakePeerGetter{peers: map[types.ID]Peer{types.ID(1): peer}}
		tr := &Transport{}
		h := newStreamHandler(tr, peerGetter, &fakeRaft{}, types.ID(2), types.ID(1))

		rw := httptest.NewRecorder()
		go h.ServeHTTP(rw, req)

		var conn *outgoingConn
		select {
		case conn = <-peer.connc:
		case <-time.After(time.Second):
			t.Fatalf("#%d: failed to attach outgoingConn", i)
		}
		if g := rw.Header().Get("X-Server-Version"); g != version.Version {
			t.Errorf("#%d: X-Server-Version = %s, want %s", i, g, version.Version)
		}
		if conn.t != tt.wtype {
			t.Errorf("#%d: type = %s, want %s", i, conn.t, tt.wtype)
		}
		conn.Close()
	}
}

// errReader implements io.Reader to facilitate a broken request.
type errReader struct{}

func (er *errReader) Read(_ []byte) (int, error) { return 0, errors.New("some error") }

type resWriterToError struct {
	code int
}

func (e *resWriterToError) Error() string                 { return "" }
func (e *resWriterToError) WriteTo(w http.ResponseWriter) { w.WriteHeader(e.code) }

type fakePeerGetter struct {
	peers map[types.ID]Peer
}

func (pg *fakePeerGetter) Get(id types.ID) Peer { return pg.peers[id] }

type fakePeer struct {
	msgs     []raftpb.Message
	snapMsgs []snap.Message
	peerURLs types.URLs
	connc    chan *outgoingConn
	paused   bool
}

func newFakePeer() *fakePeer {
	fakeURL, _ := url.Parse("http://localhost")
	return &fakePeer{
		connc:    make(chan *outgoingConn, 1),
		peerURLs: types.URLs{*fakeURL},
	}
}

func (pr *fakePeer) send(m raftpb.Message) {
	if pr.paused {
		return
	}
	pr.msgs = append(pr.msgs, m)
}

func (pr *fakePeer) sendSnap(m snap.Message) {
	if pr.paused {
		return
	}
	pr.snapMsgs = append(pr.snapMsgs, m)
}

func (pr *fakePeer) update(urls types.URLs)                { pr.peerURLs = urls }
func (pr *fakePeer) attachOutgoingConn(conn *outgoingConn) { pr.connc <- conn }
func (pr *fakePeer) activeSince() time.Time                { return time.Time{} }
func (pr *fakePeer) stop()                                 {}
func (pr *fakePeer) Pause()                                { pr.paused = true }
func (pr *fakePeer) Resume()                               { pr.paused = false }

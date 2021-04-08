package rafthttp

import (
	"context"
	"errors"
	"fmt"
	"github.com/friendlyhank/etcd-hign/net/pkg/testutil"
	"github.com/friendlyhank/etcd-hign/net/server/etcdserver/api/version"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	stats "github.com/friendlyhank/etcd-hign/net/server/etcdserver/api/v2stats"
	"go.uber.org/zap"
)

// TestStreamWriterAttachOutgoingConn tests that outgoingConn can be attached
// to streamWriter. After that, streamWriter can use it to send messages
// continuously, and closes it when stopped.
func TestStreamWriterAttachOutgoingConn(t *testing.T) {
	sw := startStreamWriter(zap.NewExample(), types.ID(0), types.ID(1), newPeerStatus(zap.NewExample(), types.ID(0), types.ID(1)), &stats.FollowerStats{}, &fakeRaft{})
	// the expected initial state of streamWriter is not working
	if _, ok := sw.writec(); ok {
		t.Errorf("initial working status = %v, want false", ok)
	}

	// repeat tests to ensure streamWriter can use last attached connection
	var wfc *fakeWriteFlushCloser
	for i := 0; i < 3; i++ {
		prevwfc := wfc
		wfc = newFakeWriteFlushCloser(nil)
		sw.attach(&outgoingConn{t: streamTypeMessage, Writer: wfc, Flusher: wfc, Closer: wfc})

		// previous attached connection should be closed
		if prevwfc != nil {
			select {
			case <-prevwfc.closed:
			case <-time.After(time.Second):
				t.Errorf("#%d: close of previous connection timed out", i)
			}
		}

		// if prevwfc != nil, the new msgc is ready since prevwfc has closed
		// if prevwfc == nil, the first connection may be pending, but the first
		// msgc is already available since it's set on calling startStreamwriter
		msgc, _ := sw.writec()
		msgc <- raftpb.Message{}

		select {
		case <-wfc.writec:
		case <-time.After(time.Second):
			t.Errorf("#%d: failed to write to the underlying connection", i)
		}
		// write chan is still available
		if _, ok := sw.writec(); !ok {
			t.Errorf("#%d: working status = %v, want true", i, ok)
		}
	}

	sw.stop()
	// write chan is unavailable since the writer is stopped.
	if _, ok := sw.writec(); ok {
		t.Errorf("working status after stop = %v, want false", ok)
	}
	if !wfc.Closed() {
		t.Errorf("failed to close the underlying connection")
	}
}

// TestStreamWriterAttachBadOutgoingConn tests that streamWriter with bad
// outgoingConn will close the outgoingConn and fall back to non-working status.
func TestStreamWriterAttachBadOutgoingConn(t *testing.T) {
	sw := startStreamWriter(zap.NewExample(), types.ID(0), types.ID(1), newPeerStatus(zap.NewExample(), types.ID(0), types.ID(1)), &stats.FollowerStats{}, &fakeRaft{})
	defer sw.stop()

	wfc := newFakeWriteFlushCloser(errors.New("blah"))
	sw.attach(&outgoingConn{t: streamTypeMessage, Writer: wfc, Flusher: wfc, Closer: wfc})

	sw.msgc <- raftpb.Message{}
	select {
	case <-wfc.closed:
	case <-time.After(time.Second):
		t.Errorf("failed to close the underlying connection in time")
	}
	// no longer working
	if _, ok := sw.writec(); ok {
		t.Errorf("working = %v, want false", ok)
	}
}

func TestStreamReaderDialRequest(t *testing.T) {
	for i, tt := range []streamType{streamTypeMessage, streamTypeMsgAppV2} {
		tr := &roundTripperRecorder{rec: &testutil.RecorderBuffered{}}
		sr := &streamReader{
			peerID: types.ID(2),
			tr:     &Transport{streamRt: tr, ClusterID: types.ID(1), ID: types.ID(1)},
			picker: mustNewURLPicker(t, []string{"http://localhost:2380"}),
			ctx:    context.Background(),
		}
		sr.dial(tt)

		act, err := tr.rec.Wait(1)
		if err != nil {
			t.Fatal(err)
		}
		req := act[0].Params[0].(*http.Request)

		wurl := fmt.Sprintf("http://localhost:2380" + tt.endpoint(zap.NewExample()) + "/1")
		if req.URL.String() != wurl {
			t.Errorf("#%d: url = %s, want %s", i, req.URL.String(), wurl)
		}
		if w := "GET"; req.Method != w {
			t.Errorf("#%d: method = %s, want %s", i, req.Method, w)
		}
		if g := req.Header.Get("X-Etcd-Cluster-ID"); g != "1" {
			t.Errorf("#%d: header X-Etcd-Cluster-ID = %s, want 1", i, g)
		}
		if g := req.Header.Get("X-Raft-To"); g != "2" {
			t.Errorf("#%d: header X-Raft-To = %s, want 2", i, g)
		}
	}
}

// TestStreamReaderDialResult tests the result of the dial func call meets the
// HTTP response received.
func TestStreamReaderDialResult(t *testing.T) {
	tests := []struct {
		code  int
		err   error
		wok   bool
		whalt bool
	}{
		{0, errors.New("blah"), false, false},
		{http.StatusOK, nil, true, false},
		{http.StatusMethodNotAllowed, nil, false, false},
		{http.StatusNotFound, nil, false, false},
		{http.StatusPreconditionFailed, nil, false, false},
		{http.StatusGone, nil, false, true},
	}
	for i, tt := range tests {
		h := http.Header{}
		h.Add("X-Server-Version", version.Version)
		tr := &respRoundTripper{
			code:   tt.code,
			header: h,
			err:    tt.err,
		}
		sr := &streamReader{
			peerID: types.ID(2),
			tr:     &Transport{streamRt: tr, ClusterID: types.ID(1)},
			picker: mustNewURLPicker(t, []string{"http://localhost:2380"}),
			errorc: make(chan error, 1),
			ctx:    context.Background(),
		}

		_, err := sr.dial(streamTypeMessage)
		if ok := err == nil; ok != tt.wok {
			t.Errorf("#%d: ok = %v, want %v", i, ok, tt.wok)
		}
		if halt := len(sr.errorc) > 0; halt != tt.whalt {
			t.Errorf("#%d: halt = %v, want %v", i, halt, tt.whalt)
		}
	}
}

type fakeWriteFlushCloser struct {
	mu      sync.Mutex
	err     error
	written int
	closed  chan struct{}
	writec  chan struct{}
}

func newFakeWriteFlushCloser(err error) *fakeWriteFlushCloser {
	return &fakeWriteFlushCloser{
		err:    err,
		closed: make(chan struct{}),
		writec: make(chan struct{}),
	}
}

func (wfc *fakeWriteFlushCloser) Write(p []byte) (n int, err error) {
	wfc.mu.Lock()
	defer wfc.mu.Unlock()
	select {
	case wfc.writec <- struct{}{}:
	default:
	}
	wfc.written += len(p)
	return len(p), wfc.err
}

func (wfc *fakeWriteFlushCloser) Flush() {}

func (wfc *fakeWriteFlushCloser) Close() error {
	close(wfc.closed)
	return wfc.err
}

func (wfc *fakeWriteFlushCloser) Written() int {
	wfc.mu.Lock()
	defer wfc.mu.Unlock()
	return wfc.written
}

func (wfc *fakeWriteFlushCloser) Closed() bool {
	select {
	case <-wfc.closed:
		return true
	default:
		return false
	}
}

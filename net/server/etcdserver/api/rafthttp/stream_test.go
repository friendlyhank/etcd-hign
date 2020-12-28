package rafthttp

import (
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

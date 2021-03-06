package rafthttp

import (
	"io"
	"net/http"
	"sync"
	"testing"

	stats "github.com/friendlyhank/etcd-hign/netmodule/server/etcdserver/api/v2stats"

	"github.com/friendlyhank/etcd-hign/raftmodule/raft/raftpb"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/netmodule/pkg/testutil"
	"github.com/friendlyhank/etcd-hign/netmodule/pkg/types"
)

// TestPipelineSend tests that pipeline could send data using roundtripper
// and increase success count in stats.
func TestPipelineSend(t *testing.T) {
	tr := &roundTripperRecorder{rec: testutil.NewRecorderStream()}
	picker := mustNewURLPicker(t, []string{"http://localhost:2380"})
	tp := &Transport{pipelineRt: tr}
	p := startTestPipeline(tp, picker)

	p.msgc <- raftpb.Message{Type: raftpb.MsgApp}
	tr.rec.Wait(1)
	p.stop()
	if p.followerStats.Counts.Success != 1 {
		t.Errorf("success = %d, want 1", p.followerStats.Counts.Success)
	}
}

type respRoundTripper struct {
	mu  sync.Mutex
	rec testutil.Recorder

	code   int
	header http.Header
	err    error
}

func newRespRoundTripper(code int, err error) *respRoundTripper {
	return &respRoundTripper{code: code, err: err}
}
func (t *respRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.rec != nil {
		t.rec.Record(testutil.Action{Name: "req", Params: []interface{}{req}})
	}
	return &http.Response{StatusCode: t.code, Header: t.header, Body: &nopReadCloser{}}, t.err
}

type roundTripperRecorder struct {
	rec testutil.Recorder
}

func (t *roundTripperRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.rec != nil {
		t.rec.Record(testutil.Action{Name: "req", Params: []interface{}{req}})
	}
	return &http.Response{StatusCode: http.StatusNoContent, Body: &nopReadCloser{}}, nil
}

type nopReadCloser struct{}

func (n *nopReadCloser) Read(p []byte) (int, error) { return 0, io.EOF }
func (n *nopReadCloser) Close() error               { return nil }

func startTestPipeline(tr *Transport, picker *urlPicker) *pipeline {
	p := &pipeline{
		peerID:        types.ID(1),
		tr:            tr,
		picker:        picker,
		status:        newPeerStatus(zap.NewExample(), tr.ID, types.ID(1)),
		raft:          &fakeRaft{},
		followerStats: &stats.FollowerStats{},
		errorc:        make(chan error, 1),
	}
	p.start()
	return p
}

package rafthttp

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/net/pkg/pbutil"
	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"
	"github.com/friendlyhank/etcd-hign/net/server/etcdserver/api/version"
)

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

// errReader implements io.Reader to facilitate a broken request.
type errReader struct{}

func (er *errReader) Read(_ []byte) (int, error) { return 0, errors.New("some error") }

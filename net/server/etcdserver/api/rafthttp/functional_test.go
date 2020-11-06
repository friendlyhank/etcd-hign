package rafthttp

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
)

func TestSendMessage(t *testing.T) {
	//member 1
	tr := &Transport{
		ID:        types.ID(1),
		ClusterID: types.ID(1),
	}
	tr.Start()
	//httptest测试包测试
	srv := httptest.NewServer(tr.Handler())
	defer srv.Close()

	//member 2
	tr2 := &Transport{
		ID:        types.ID(2),
		ClusterID: types.ID(1),
	}
	tr2.Start()
	srv2 := httptest.NewServer(tr2.Handler())
	defer srv2.Close()

	tr.AddPeer(types.ID(2), []string{srv2.URL})
	defer tr.Stop()
	tr2.AddPeer(types.ID(1), []string{srv.URL})
	defer tr2.Stop()
	if !waitStreamWorking(tr.Get(types.ID(2)).(*peer)) {
		t.Fatalf("stream from 1 to 2 is not in work as expected")
	}
}

func waitStreamWorking(p *peer) bool {
	for i := 0; i < 1000; i++ {
		time.Sleep(time.Millisecond)
		if _, ok := p.msgAppV2Writer.writec(); !ok {
			continue
		}
		if _, ok := p.writer.writec(); !ok {
			continue
		}
		return true
	}
	return false
}

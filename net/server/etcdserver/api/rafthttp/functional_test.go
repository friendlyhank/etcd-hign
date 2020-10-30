package rafthttp

import (
	"net/http/httptest"
	"testing"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
)

func TestSendMessage(t *testing.T) {
	//member 1
	tr := &Transport{
		ID:        types.ID(1),
		ClusterID: types.ID(1),
	}
	tr.Start()
	srv := httptest.NewServer(tr.Handler())
	defer srv.Close()

	//member 2
	tr2 := &Transport{
		ID:        types.ID(1),
		ClusterID: types.ID(2),
	}
	tr2.Start()
	srv2 := httptest.NewServer(tr.Handler())
	defer srv2.Close()

	tr.AddPeer(types.ID(2), []string{srv2.URL})
	tr2.AddPeer(types.ID(1), []string{srv.URL})
}

package etcdhttp

import (
	"net/http"

	"github.com/friendlyhank/etcd-hign/net/server/etcdserver/api/rafthttp"

	"github.com/friendlyhank/etcd-hign/net/server/etcdserver"
	"go.uber.org/zap"
)

func NewPeerHandler(lg *zap.Logger, s etcdserver.ServerPeerV2) http.Handler {
	return newPeerHandler(lg, s, s.RaftHandler(), nil, nil)
}

func newPeerHandler(
	lg *zap.Logger,
	s etcdserver.Server,
	raftHandler http.Handler,
	leaseHandler http.Handler,
	hashKVHandler http.Handler,
) http.Handler {
	if lg == nil {
		lg = zap.NewNop()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", http.NotFound)
	mux.Handle(rafthttp.RaftPrefix, raftHandler)
	mux.Handle(rafthttp.RaftPrefix+"/", raftHandler)
	return mux
}

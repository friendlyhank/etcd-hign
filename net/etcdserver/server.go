package etcdserver

import (
	"net/http"

	"github.com/friendlyhank/etcd-hign/net/etcdserver/api/rafthttp"
)

type ServerPeer interface {
}

type ServerV3 interface {
}

type ServerV2 interface {
}

type Server interface {
}

type EtcdServer struct {
	r raftNode
}

func NewServer(cfg ServerConfig) (srv *EtcdServer, err error) {
	srv = &EtcdServer{}

	tr := &rafthttp.Transport{}
	srv.r.transport = tr
	return srv, nil
}

func (s *EtcdServer) RaftHandler() http.Handler { return s.r.transport.Handler() }

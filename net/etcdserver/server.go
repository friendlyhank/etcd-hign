package etcdserver

import (
	"net/http"

	"github.com/friendlyhank/etcd-hign/net/raft"

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

	var (
		n raft.Node
	)

	//启动node
	n = startNode()

	srv = &EtcdServer{
		r: *newRaftNode(raftNodeConfig{
			Node: n, //这货隐藏的比较深
		}),
	}

	tr := &rafthttp.Transport{}
	srv.r.transport = tr
	return srv, nil
}

func (s *EtcdServer) Start() {
	s.start()
}

func (s *EtcdServer) start() {
	go s.run()
}

type raftReadyHandler struct {
}

func (s *EtcdServer) run() {
	rh := &raftReadyHandler{}
	s.r.start(rh)
}

func (s *EtcdServer) RaftHandler() http.Handler { return s.r.transport.Handler() }

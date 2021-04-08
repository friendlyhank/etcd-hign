package etcdserver

import (
	"net/http"

	stats "github.com/friendlyhank/etcd-hign/net/server/etcdserver/api/v2stats"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"github.com/friendlyhank/etcd-hign/net/server/etcdserver/api/membership"

	"github.com/friendlyhank/etcd-hign/net/raft"

	"github.com/friendlyhank/etcd-hign/net/server/etcdserver/api/rafthttp"
)

type ServerPeer interface {
	RaftHandler() http.Handler
}

type ServerV3 interface {
}

type ServerV2 interface {
}

type Server interface {
}

type EtcdServer struct {
	r raftNode

	lg *zap.Logger

	errorc chan error
}

func NewServer(cfg ServerConfig) (srv *EtcdServer, err error) {

	var (
		n  raft.Node
		id types.ID //当前节点唯一id
		cl *membership.RaftCluster
	)

	var (
		remotes []*membership.Member
	)

	//根据urlmap设置集群信息和member信息
	cl, err = membership.NewClusterFromURLsMap(nil, cfg.InitialClusterToken, cfg.InitialPeerURLsMap)
	//启动node
	id, n = startNode(cfg, cl)

	lstats := stats.NewLeaderStats(cfg.Logger, id.String())

	srv = &EtcdServer{
		lg:     cfg.Logger,
		errorc: make(chan error, 1),
		r: *newRaftNode(raftNodeConfig{
			Node: n, //这货隐藏的比较深
		}),
	}
	// TODO: move transport initialization near the definition of remote
	tr := &rafthttp.Transport{
		Logger:      cfg.Logger,
		TLSInfo:     cfg.PeerTLSInfo,
		ID:          id,
		LeaderStats: lstats,
	}
	//启动etcd核心网络传输组件
	if err = tr.Start(); err != nil {
		return nil, err
	}

	//remotes为排除当前节点的所有节点
	for _, m := range remotes {
		if m.ID != id {
			tr.AddRemote(m.ID, m.PeerURLs)
		}
	}

	//addPeer 会startPeer并且启动监听
	for _, m := range cl.Members() {
		if m.ID != id {
			tr.AddPeer(m.ID, m.PeerURLs)
		}
	}
	srv.r.transport = tr

	//设置节点信息
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

//设置EtcdServer的三大Handler RaftHandler() LeaseHandler()
func (s *EtcdServer) RaftHandler() http.Handler { return s.r.transport.Handler() }

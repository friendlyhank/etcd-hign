package etcdserver

import (
	"github.com/friendlyhank/etcd-hign/netmodule/pkg/types"
	"github.com/friendlyhank/etcd-hign/netmodule/server/etcdserver/api/membership"
	"github.com/friendlyhank/etcd-hign/netmodule/server/etcdserver/api/rafthttp"
	"github.com/friendlyhank/etcd-hign/raftmodule/raft"
)

type raftNode struct {
	raftNodeConfig //这里隐藏很多重要信息 transport,Node
}

type raftNodeConfig struct {
	raft.Node
	transport rafthttp.Transporter
}

func newRaftNode(cfg raftNodeConfig) *raftNode {
	r := &raftNode{
		raftNodeConfig: cfg,
	}
	return r
}

func (r *raftNode) start(rh *raftReadyHandler) {
	go func() {
		for {
			select {
			case rd := <-r.Ready():
				msg := rd.Messages
				r.transport.Send(msg)
			}
		}
	}()
}

func startNode(cfg ServerConfig, cl *membership.RaftCluster) (id types.ID, n raft.Node) {
	member := cl.MemberByName(cfg.Name)
	id = member.ID
	n = raft.StartNode()
	return id, n
}

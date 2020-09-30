package etcdserver

import (
	"github.com/friendlyhank/etcd-hign/net/etcdserver/api/rafthttp"
	"github.com/friendlyhank/etcd-hign/net/raft"
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
			case <-r.Ready():

			}
		}
	}()
}

func startNode() (n raft.Node) {
	n = raft.StartNode()
	return n
}

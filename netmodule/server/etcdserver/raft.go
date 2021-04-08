package etcdserver

import (
	"time"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/netmodule/pkg/types"
	"github.com/friendlyhank/etcd-hign/netmodule/server/etcdserver/api/membership"
	"github.com/friendlyhank/etcd-hign/netmodule/server/etcdserver/api/rafthttp"
	"github.com/friendlyhank/etcd-hign/raftmodule/raft"
)

type raftNode struct {
	lg *zap.Logger

	raftNodeConfig //这里隐藏很多重要信息 transport,Node

	// utility
	ticker *time.Ticker //选举用的定时 ticker
}

type raftNodeConfig struct {
	raft.Node
	heartbeat time.Duration //心跳消息用于raft定时选举 for logging
	// transport specifies the transport to send and receive msgs to members.
	// Sending messages MUST NOT block. It is okay to drop messages, since
	// clients should timeout and reissue their messages.
	// If transport is nil, server will panic.
	//transport可以发送和接收消息
	transport rafthttp.Transporter
}

func newRaftNode(cfg raftNodeConfig) *raftNode {
	r := &raftNode{
		raftNodeConfig: cfg,
	}
	//设置选举定时
	if r.heartbeat == 0 {
		r.ticker = &time.Ticker{}
	} else {
		r.ticker = time.NewTicker(r.heartbeat)
	}
	return r
}

func (r *raftNode) tick() {
	r.Tick()
}

//raft调用transport去发送消息
func (r *raftNode) start(rh *raftReadyHandler) {
	go func() {
		for {
			select {
			case <-r.ticker.C:
				r.tick()
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

	//raft配置相关
	c := &raft.Config{
		ID: uint64(id),
	}

	n = raft.StartNode(c)
	return id, n
}

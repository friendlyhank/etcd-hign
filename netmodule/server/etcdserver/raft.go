package etcdserver

import (
	"encoding/json"
	"time"

	"github.com/friendlyhank/etcd-hign/raftmodule/raft/raftpb"

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
		islead := false
		for {
			select {
			case <-r.ticker.C: //定时选举
				r.tick()
			// the leader can write to its disk in parallel with replicating to the followers and them
			// writing to their disks.
			// For more details, check raft thesis 10.2.1
			case rd := <-r.Ready(): //处在ready状态,然后发送消息
				//如果是领导者
				if islead {
					r.transport.Send(r.processMessages(rd.Messages))
				}

				if !islead {
					// finish processing incoming messages before we signal raftdone chan
					msgs := r.processMessages(rd.Messages)

					// gofail: var raftBeforeFollowerSend struct{}
					r.transport.Send(msgs)
				}

				//通知发送下一批消息
				r.Advance()
			}
		}
	}()
}

func (r *raftNode) processMessages(ms []raftpb.Message) []raftpb.Message {
	for i := len(ms) - 1; i >= 0; i-- {
	}
	return ms
}

func startNode(cfg ServerConfig, cl *membership.RaftCluster, ids []types.ID) (id types.ID, n raft.Node) {
	var err error
	member := cl.MemberByName(cfg.Name)

	peers := make([]raft.Peer, len(ids))
	for i, id := range ids {
		var ctx []byte
		ctx, err = json.Marshal((*cl).Member(id))
		if err != nil {
			cfg.Logger.Panic("failed to marshal member", zap.Error(err))
		}
		peers[i] = raft.Peer{ID: uint64(id), Context: ctx}
	}
	id = member.ID
	cfg.Logger.Info(
		"starting local member",
		zap.String("local-member-id", id.String()),
		zap.String("cluster-id", cl.ID().String()),
	)

	//raft配置相关
	c := &raft.Config{
		ID:            uint64(id),
		ElectionTick:  cfg.ElectionTicks,
		HeartbeatTick: 1,
		CheckQuorum:   true,
		PreVote:       cfg.PreVote,
	}

	n = raft.StartNode(c, peers)
	return id, n
}

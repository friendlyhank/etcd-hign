package raft

import (
	"context"

	pb "github.com/friendlyhank/etcd-hign/raftmodule/raft/raftpb"
)

type SnapshotStatus int

// Ready encapsulates the entries and messages that are ready to read,
// be saved to stable storage, committed or sent to other peers.
// All fields in Ready are read-only.
//Ready用于发送条目和消息
//会被存储在stable storage中,提交或者发送给其他peers
//Ready的所有字段都是只读的
type Ready struct {
	Messages []pb.Message
}

type Node interface {
	// Tick increments the internal logical clock for the Node by a single tick. Election
	// timeouts and heartbeat timeouts are in units of ticks.
	//用于定时选举
	Tick()

	// Step advances the state machine using the given message. ctx.Err() will be returned, if any.
	Step(ctx context.Context, msg pb.Message) error

	// Ready returns a channel that returns the current point-in-time state.
	// Users of the Node must call Advance after retrieving the state returned by Ready.
	//
	// NOTE: No committed entries from the next Ready may be applied until all committed entries
	// and snapshots from the previous one have finished.
	Ready() <-chan Ready //准备就绪可发送消息

	// ReportUnreachable reports the given node is not reachable for the last send.
	//报告网络服务的不可用
	ReportUnreachable(id uint64)
}

type Peer struct {
	ID      uint64 //节点信息
	Context []byte
}

// StartNode returns a new Node given configuration and a list of raft peers.
// It appends a ConfChangeAddNode entry for each given peer to the initial log.
//
// Peers must not be zero length; call RestartNode in that case.
//启动node
func StartNode(c *Config, peers []Peer) Node {
	rn, err := NewRawNode(c)
	if err != nil {
		panic(err)
	}

	//初始化成为跟随者信息同时初始化投票相关信息
	rn.Bootstrap(peers)

	n := newNode(rn)

	//这里会去发送消息
	go n.run()
	return &n
}

// node is the canonical implementation of the Node interface
//节点信息,这个作为ectd的重要点
type node struct {
	recvc    chan pb.Message //接收消息
	readyc   chan Ready      //发送消息就绪状态
	advancec chan struct{}   //发送消息readyc开关控制器
	tickc    chan struct{}   //竞选领导者的定时
	rn       *RawNode
}

func newNode(rn *RawNode) node {
	return node{
		recvc:    make(chan pb.Message),
		readyc:   make(chan Ready),
		advancec: make(chan struct{}),
		// make tickc a buffered chan, so raft node can buffer some ticks when the node
		// is busy processing raft messages. Raft node will resume process buffered
		// ticks when it becomes idle.
		tickc: make(chan struct{}, 128),
		rn:    rn,
	}
}

func (n *node) run() {
	var readyc chan Ready
	var advancec chan struct{}
	var rd Ready

	r := n.rn.raft

	for {
		if advancec != nil {
			readyc = nil
		} else if n.rn.HasReady() {
			rd = n.rn.readyWithoutAccept()
			readyc = n.readyc
		}
		select {
		case m := <-n.recvc: //接收消息
			// filter out response message from unknown From.
			if pr := r.prs.Progress[m.From]; pr != nil {
				r.Step(m) //把接收到的消息交给raft处理
			}
		case <-n.tickc: //启动定时
			n.rn.Tick()
		case readyc <- rd: //如果有消息，写入到n.readyc
			advancec = n.advancec
		case <-advancec:
			rd = Ready{}
			advancec = nil
		}
	}
}

// Tick increments the internal logical clock for this Node. Election timeouts
// and heartbeat timeouts are in units of ticks.
//启动节点的定时选举
func (n *node) Tick() {
	select {
	case n.tickc <- struct{}{}:
	default:
	}
}

func (n *node) Step(ctx context.Context, m pb.Message) error {
	return n.step(ctx, m)
}

func (n *node) step(ctx context.Context, m pb.Message) error {
	return nil
}

// newReady- 在这里去new Ready
func newReady(r *raft) Ready {
	rd := Ready{
		Messages: r.msgs,
	}
	return rd
}

func (n *node) Ready() <-chan Ready { return n.readyc }

func (n *node) ReportUnreachable(id uint64) {
	select {
	case n.recvc <- pb.Message{Type: pb.MsgUnreachable, From: id}:
	}
}

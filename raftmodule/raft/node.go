package raft

import (
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

	// Ready returns a channel that returns the current point-in-time state.
	// Users of the Node must call Advance after retrieving the state returned by Ready.
	//
	// NOTE: No committed entries from the next Ready may be applied until all committed entries
	// and snapshots from the previous one have finished.
	Ready() <-chan Ready //准备就绪可发送消息
}

type Peer struct {
}

//启动node
func StartNode(c *Config) Node {
	rn, err := NewRawNode()
	if err != nil {
		panic(err)
	}
	n := newNode(rn)

	//这里会去发送消息
	go n.run()
	return &n
}

// node is the canonical implementation of the Node interface
//节点信息,这个作为ectd的重要点
type node struct {
	readyc chan Ready
	tickc  chan struct{}
	rn     *RawNode
}

func newNode(rn *RawNode) node {
	return node{
		readyc: make(chan Ready),
		// make tickc a buffered chan, so raft node can buffer some ticks when the node
		// is busy processing raft messages. Raft node will resume process buffered
		// ticks when it becomes idle.
		tickc: make(chan struct{}, 128),
		rn:    rn,
	}
}

func (n *node) run() {
	var readyc chan Ready
	var rd Ready
	for {
		//从这里去写入消息到channel,然后channel接收端会不断循环发送消息
		rd = n.rn.readyWithoutAccept()
		readyc = n.readyc

		select {
		case <-n.tickc:
			n.rn.Tick()
		case readyc <- rd:
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

// newReady- 在这里去new Ready
func newReady(r *raft) Ready {
	rd := Ready{
		Messages: r.msgs,
	}
	return rd
}

func (n *node) Ready() <-chan Ready { return n.readyc }
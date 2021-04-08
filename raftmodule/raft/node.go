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
	Ready() <-chan Ready
}

type node struct {
	readyc chan Ready

	rn *RawNode
}

//启动node
func StartNode() Node {
	rn, err := NewRawNode()
	if err != nil {
		panic(err)
	}
	n := newNode(rn)

	//这里会去发送消息
	go n.run()
	return &n
}

func newNode(rn *RawNode) node {
	return node{
		readyc: make(chan Ready),
		rn:     rn,
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
		case readyc <- rd:
		}
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

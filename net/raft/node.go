package raft

import (
	pb "github.com/friendlyhank/etcd-hign/net/v3/raft/raftpb"
)

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

func StartNode() Node {
	rn, err := NewRawNode()
	if err != nil {
		panic(err)
	}
	n := newNode(rn)

	//启动node
	//Ready在这里
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

	//TODO HANK 写死消息
	rd = Ready{
		Messages: []pb.Message{
			pb.Message{
				Type: pb.MsgVote,
				To:   1849879258734672239,
				From: 13803658152347727308,
			},
		},
	}

	for {
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

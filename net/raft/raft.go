package raft

import pb "github.com/friendlyhank/etcd-hign/net/v3/raft/raftpb"

// None is a placeholder node ID used when there is no leader.
const None uint64 = 0

type raft struct {
	msgs []pb.Message
}

func newRaft() *raft {
	r := &raft{}
	return r
}

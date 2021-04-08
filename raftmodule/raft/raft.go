package raft

import pb "github.com/friendlyhank/etcd-hign/raftmodule/raft/raftpb"

// None is a placeholder node ID used when there is no leader.
const None uint64 = 0

// Config contains the parameters to start a raft.
//raft相关
type Config struct {
	// ID is the identity of the local raft. ID cannot be 0.
	ID uint64
}

type raft struct {
	msgs []pb.Message

	// the leader id
	//领导者id
	lead uint64

	tick func() //选举时候需要定时执行的方法
}

func newRaft() *raft {
	r := &raft{}
	return r
}

//定时选举
func (r *raft) tickElection() {

}

func (r *raft) becomeFollower(term uint64, lead uint64) {
	r.tick = r.tickElection
}

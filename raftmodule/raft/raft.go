package raft

import (
	pb "github.com/friendlyhank/etcd-hign/raftmodule/raft/raftpb"
)

// None is a placeholder node ID used when there is no leader.
const None uint64 = 0

// Possible values for StateType.
const (
	StateFollower StateType = iota
	StateCandidate
	StateLeader
	StatePreCandidate
)

// Possible values for CampaignType
const (
	// campaignPreElection represents the first phase of a normal election when
	// Config.PreVote is true.
	campaignPreElection CampaignType = "CampaignPreElection"
	// campaignElection represents a normal (time-based) election (the second phase
	// of the election when Config.PreVote is true).
	campaignElection CampaignType = "CampaignElection"
)

// CampaignType represents the type of campaigning
// the reason we use the type of string instead of uint64
// is because it's simpler to compare and fill in raft entries
type CampaignType string //候选者类型 候选者和预候选者

// StateType represents the role of a node in a cluster.
type StateType uint64

// Config contains the parameters to start a raft.
//raft相关
type Config struct {
	// ID is the identity of the local raft. ID cannot be 0.
	ID uint64

	// PreVote enables the Pre-Vote algorithm described in raft thesis section
	// 9.6. This prevents disruption when a node that has been partitioned away
	// rejoins the cluster.
	PreVote bool
}

type raft struct {
	id uint64

	Term uint64 //任期号
	Vote uint64 //投票id号

	state StateType

	msgs []pb.Message

	// the leader id
	//领导者id
	lead uint64

	preVote bool

	tick func() //选举时候需要定时执行的方法

	logger Logger
}

func newRaft() *raft {
	r := &raft{}
	return r
}

// maybeCommit attempts to advance the commit index. Returns true if
// the commit index changed (in which case the caller should call
// r.bcastAppend).
func (r *raft) reset(term uint64) {
	if r.Term != term {
		r.Term = term
	}
	r.lead = None
}

//定时选举
func (r *raft) tickElection() {
	r.Step(pb.Message{From: r.id, Type: pb.MsgHup})
}

// tickHeartbeat is run by leaders to send a MsgBeat after r.heartbeatTimeout.
//成为领导者后发送心跳
func (r *raft) tickHeartbeat() {

}

func (r *raft) becomeFollower(term uint64, lead uint64) {
	r.reset(term)
	r.tick = r.tickElection
	r.lead = lead
	r.state = StateFollower
}

func (r *raft) becomeCandidate() {
	// TODO(xiangli) remove the panic when the raft implementation is stable
	if r.state == StateLeader {
		panic("invalid transition [leader -> candidate]")
	}
	r.reset(r.Term + 1)
	r.tick = r.tickElection
	r.Vote = r.id
	r.state = StateCandidate
}

func (r *raft) becomePreCandidate() {
	// TODO(xiangli) remove the panic when the raft implementation is stable
	if r.state == StateLeader {
		panic("invalid transition [leader -> pre-candidate]")
	}
	// Becoming a pre-candidate changes our step functions and state,
	// but doesn't change anything else. In particular it does not increase
	// r.Term or change r.Vote.
	r.tick = r.tickElection
	r.lead = None
	r.state = StatePreCandidate
}

func (r *raft) becomeLeader() {
	// TODO(xiangli) remove the panic when the raft implementation is stable
	if r.state == StateFollower {
		panic("invalid transition [follower -> leader]")
	}
	r.reset(r.Term)
	r.tick = r.tickHeartbeat
	r.lead = r.id
	r.state = StateLeader
}

//晋升为候选人或预候选人
func (r *raft) hup(t CampaignType) {
	r.campaign(t)
}

// campaign transitions the raft instance to candidate state. This must only be
// called after verifying that this is a legitimate transition.
func (r *raft) campaign(t CampaignType) {
	if t == campaignPreElection {

	}
}

func (r *raft) Step(m pb.Message) error {
	switch m.Type {
	case pb.MsgHup:
		if r.preVote {

		} else {

		}
	}
	return nil
}

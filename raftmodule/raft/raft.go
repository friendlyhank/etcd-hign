package raft

import (
	"github.com/friendlyhank/etcd-hign/raftmodule/raft/quorum"
	pb "github.com/friendlyhank/etcd-hign/raftmodule/raft/raftpb"
	"github.com/friendlyhank/etcd-hign/raftmodule/raft/tracker"
)

// None is a placeholder node ID used when there is no leader.
const None uint64 = 0

// Possible values for StateType.
//raft竞选的状态
const (
	StateFollower StateType = iota
	StateCandidate
	StateLeader
	StatePreCandidate
)

// Possible values for CampaignType
//候选者类型 候选者和预候选者
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
type CampaignType string

// StateType represents the role of a node in a cluster.
type StateType uint64

// Config contains the parameters to start a raft.
//raft配置相关
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

	// TODO(tbg): rename to trk.
	prs tracker.ProgressTracker //投票相关统计

	state StateType //竞选状态信息

	msgs []pb.Message

	// the leader id
	//领导者id
	lead uint64

	preVote bool //是否需要预候选人

	tick func()   //选举时候需要定时执行的方法
	step stepFunc //竞选的下一个步骤

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

//初始化成为跟随着信息
func (r *raft) becomeFollower(term uint64, lead uint64) {
	r.step = stepFollower
	r.reset(term)
	r.tick = r.tickElection
	r.lead = lead
	r.state = StateFollower
	r.logger.Infof("%x became follower at term %d", r.id, r.Term)
}

//初始化成为候选人信息
func (r *raft) becomeCandidate() {
	// TODO(xiangli) remove the panic when the raft implementation is stable
	if r.state == StateLeader {
		panic("invalid transition [leader -> candidate]")
	}
	r.step = stepCandidate
	r.reset(r.Term + 1)
	r.tick = r.tickElection
	r.Vote = r.id
	r.state = StateCandidate
	r.logger.Infof("%x became candidate at term %d", r.id, r.Term)
}

//初始化成为预候选人信息
func (r *raft) becomePreCandidate() {
	// TODO(xiangli) remove the panic when the raft implementation is stable
	if r.state == StateLeader {
		panic("invalid transition [leader -> pre-candidate]")
	}
	// Becoming a pre-candidate changes our step functions and state,
	// but doesn't change anything else. In particular it does not increase
	// r.Term or change r.Vote.
	r.step = stepCandidate
	r.tick = r.tickElection
	r.lead = None
	r.state = StatePreCandidate
	r.logger.Infof("%x became pre-candidate at term %d", r.id, r.Term)
}

//初始化成为领导者信息
func (r *raft) becomeLeader() {
	// TODO(xiangli) remove the panic when the raft implementation is stable
	if r.state == StateFollower {
		panic("invalid transition [follower -> leader]")
	}
	r.step = stepLeader
	r.reset(r.Term)
	r.tick = r.tickHeartbeat
	r.lead = r.id
	r.state = StateLeader
	r.logger.Infof("%x became leader at term %d", r.id, r.Term)
}

//晋升为候选人或预候选人
func (r *raft) hup(t CampaignType) {
	if r.state == StateLeader {

	}
	r.campaign(t)
}

// campaign transitions the raft instance to candidate state. This must only be
// called after verifying that this is a legitimate transition.
//晋升为候选人或预候选人
func (r *raft) campaign(t CampaignType) {
	var term uint64
	var voteMsg pb.MessageType
	if t == campaignPreElection {
		r.becomePreCandidate()
		voteMsg = pb.MsgPreVote //准备发送预选投票消息
	} else {
		r.becomeCandidate()
		voteMsg = pb.MsgVote //准备发送投票消息
		term = r.Term
	}
}

//接收投票统计
func (r *raft) poll(id uint64, t pb.MessageType, v bool) (granted int, rejected int, result quorum.VoteResult) {
	return
}

//执行竞选的状态
func (r *raft) Step(m pb.Message) error {
	switch m.Type {
	case pb.MsgHup: //晋升成为候选人
		if r.preVote {
			r.hup(campaignPreElection)
		} else {
			r.hup(campaignElection)
		}

	default:
		err := r.step(r, m)
		if err != nil {
			return err
		}
	}
	return nil
}

type stepFunc func(r *raft, m pb.Message) error

func stepLeader(r *raft, m pb.Message) error {
	return nil
}

// stepCandidate is shared by StateCandidate and StatePreCandidate; the difference is
// whether they respond to MsgVoteResp or MsgPreVoteResp.
func stepCandidate(r *raft, m pb.Message) error {
	// Only handle vote responses corresponding to our candidacy (while in
	// StateCandidate, we may get stale MsgPreVoteResp messages in this term from
	// our pre-candidate state).
	return nil
}

func stepFollower(r *raft, m pb.Message) error {
	return nil
}

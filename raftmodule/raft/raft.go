package raft

import (
	"sort"

	"github.com/friendlyhank/etcd-hign/raftmodule/raft/confchange"

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

	// Logger is the logger used for raft log. For multinode which can host
	// multiple raft group, each raft group can have its own logger
	Logger Logger
}

func (c *Config) validate() error {
	if c.Logger == nil {
		c.Logger = raftLogger
	}
	return nil
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

func newRaft(c *Config) *raft {
	if err := c.validate(); err != nil {
		panic(err.Error())
	}
	r := &raft{
		logger: c.Logger,
		prs:    tracker.MakeProgressTracker(0),
	}
	return r
}

// send persists state to stable storage and then sends to its mailbox.
//准备要发送的消息体
func (r *raft) send(m pb.Message) {
	r.msgs = append(r.msgs, m)
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
	//给自己投一票
	if _, _, res := r.poll(r.id, voteRespMsgType(voteMsg), true); res == quorum.VoteWon {
		//每个节点投票的时间不一样,所以在有可能别的节点已经投票自己再投一票就胜出的情况
		// We won the election after voting for ourselves (which must mean that
		// this is a single-node cluster). Advance to the next state.
		//预候选人成为候选人
		if t == campaignPreElection {
			r.campaign(campaignElection)
		} else {
			r.becomeLeader() //成为领导者
		}
		return
	}

	var ids []uint64
	{
		idMap := r.prs.Voters.IDs()
		ids = make([]uint64, 0, len(idMap))
		for id := range idMap {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	}
	//向各个节点发送投票请求
	for _, id := range ids {
		if id == r.id {
			continue
		}
		r.send(pb.Message{Term: term, To: id, Type: voteMsg})
	}
}

//接收投票统计
func (r *raft) poll(id uint64, t pb.MessageType, v bool) (granted int, rejected int, result quorum.VoteResult) {
	if v {
		r.logger.Infof("%x received %s from %x at term %d", r.id, t, id, r.Term)
	} else {
		r.logger.Infof("%x received %s rejection from %x at term %d", r.id, t, id, r.Term)
	}
	//记录票数
	r.prs.RecordVote(id, v)
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

//申请修改配置信息
func (r *raft) applyConfChange(cc pb.ConfChangeV2) pb.ConfState {
	cfg, prs, err := func() (tracker.Config, tracker.ProgressMap, error) {
		changer := confchange.Changer{
			Tracker: r.prs,
		}
		return changer.Simple(cc.Changes...)
	}()
	if err != nil {
		// TODO(tbg): return the error to the caller.
		panic(err)
	}
	return r.switchToConfig(cfg, prs)
}

// switchToConfig reconfigures this node to use the provided configuration. It
// updates the in-memory state and, when necessary, carries out additional
// actions such as reacting to the removal of nodes or changed quorum
// requirements.
//
// The inputs usually result from restoring a ConfState or applying a ConfChange.
//让更改的raft配置信息并生效
func (r *raft) switchToConfig(cfg tracker.Config, prs tracker.ProgressMap) pb.ConfState {
	//复制新的配置信息
	r.prs.Config = cfg
	r.prs.Progress = prs
	return pb.ConfState{}
}

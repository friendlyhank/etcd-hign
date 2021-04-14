package raft

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

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

// lockedRand is a small wrapper around rand.Rand to provide
// synchronization among multiple raft groups. Only the methods needed
// by the code are exposed (e.g. Intn).
type lockedRand struct {
	mu   sync.Mutex
	rand *rand.Rand
}

func (r *lockedRand) Intn(n int) int {
	r.mu.Lock()
	v := r.rand.Intn(n)
	r.mu.Unlock()
	return v
}

var globalRand = &lockedRand{
	rand: rand.New(rand.NewSource(time.Now().UnixNano())),
}

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

	// ElectionTick is the number of Node.Tick invocations that must pass between
	// elections. That is, if a follower does not receive any message from the
	// leader of current term before ElectionTick has elapsed, it will become
	// candidate and start an election. ElectionTick must be greater than
	// HeartbeatTick. We suggest ElectionTick = 10 * HeartbeatTick to avoid
	// unnecessary leader switching.
	ElectionTick int //TODO HANK 重点研究下
	// HeartbeatTick is the number of Node.Tick invocations that must pass between
	// heartbeats. That is, a leader sends heartbeat messages to maintain its
	// leadership every HeartbeatTick ticks.
	HeartbeatTick int //TODO HANK 重点研究下

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
	Vote uint64 //当前节点已经投票的id

	// TODO(tbg): rename to trk.
	prs tracker.ProgressTracker //投票相关统计

	state StateType //竞选状态信息

	msgs []pb.Message

	// the leader id
	//领导者id
	lead uint64

	// number of ticks since it reached last electionTimeout when it is leader
	// or candidate.
	// number of ticks since it reached last electionTimeout or received a
	// valid message from current leader when it is a follower.
	electionElapsed int //选举时间摆钟,累加

	// number of ticks since it reached last heartbeatTimeout.
	// only leader keeps heartbeatElapsed.
	heartbeatElapsed int //心跳时间摆钟,累加

	preVote bool //是否需要预候选人

	electionTimeout  int //选举超时,如果超时会触发新一轮选举
	heartbeatTimeout int //心跳超时

	checkQuorum bool //leader检查是否大多数节点处于活跃状态

	// randomizedElectionTimeout is a random number between
	// [electiontimeout, 2 * electiontimeout - 1]. It gets reset
	// when raft changes its state to follower or candidate.
	//带随机的选举超时,取值范围在[electiontimeout, 2 * electiontimeout - 1]
	randomizedElectionTimeout int

	tick func()   //选举时候需要定时执行的方法
	step stepFunc //竞选的下一个步骤

	logger Logger
}

func newRaft(c *Config) *raft {
	if err := c.validate(); err != nil {
		panic(err.Error())
	}
	r := &raft{
		id:               c.ID,
		logger:           c.Logger,
		prs:              tracker.MakeProgressTracker(0),
		electionTimeout:  c.ElectionTick,
		heartbeatTimeout: c.HeartbeatTick,
	}
	return r
}

// send persists state to stable storage and then sends to its mailbox.
//准备要发送的消息体
func (r *raft) send(m pb.Message) {
	if m.From == None {
		m.From = r.id
	}
	//生成消息和发送消息逻辑分开，使用append消息,一次可以发送多个消息体
	r.msgs = append(r.msgs, m)
}

// sendHeartbeat sends a heartbeat RPC to the given peer.
func (r *raft) sendHeartbeat(to uint64, ctx []byte) {
	m := pb.Message{
		To:   to,
		Type: pb.MsgHeartbeat,
	}
	r.send(m)
}

// bcastHeartbeat sends RPC, without entries to all the peers.
//广播心跳消息
func (r *raft) bcastHeartbeat() {
	r.bcastHeartbeatWithCtx(nil)
}

func (r *raft) bcastHeartbeatWithCtx(ctx []byte) {
	r.prs.Visit(func(id uint64, _ *tracker.Progress) {
		if id == r.id {
			return
		}
		r.sendHeartbeat(id, ctx)
	})
}

// maybeCommit attempts to advance the commit index. Returns true if
// the commit index changed (in which case the caller should call
// r.bcastAppend).
func (r *raft) reset(term uint64) {
	if r.Term != term {
		r.Term = term
		r.Vote = None
	}
	r.lead = None
	r.electionElapsed = 0  //选举时间摆钟变为0
	r.heartbeatElapsed = 0 //心跳时间摆钟变为0
	r.resetRandomizedElectionTimeout()
}

//定时选举
func (r *raft) tickElection() {
	r.electionElapsed++ //每次定时选举的时间摆钟加一
	if r.pastElectionTimeout() {
		r.electionElapsed = 0
		r.Step(pb.Message{From: r.id, Type: pb.MsgHup})
	}
}

// tickHeartbeat is run by leaders to send a MsgBeat after r.heartbeatTimeout.
//成为领导者后发送心跳
func (r *raft) tickHeartbeat() {
	r.heartbeatElapsed++
	r.electionElapsed++ //tickElection变为tickHeartbeat,所以还需要统计选举超时状况

	//选举时间摆钟大于选举超时的时候,leader会自查自身健康状况,检查发送心跳的节点是否还超过半数
	//如果小于半数,说明leader节点接收follower心跳答复出现问题
	if r.electionElapsed >= r.electionTimeout {
		r.electionElapsed = 0
		if r.checkQuorum {
			r.Step(pb.Message{From: r.id, Type: pb.MsgCheckQuorum})
		}
	}

	if r.state != StateLeader {
		return
	}

	if r.heartbeatElapsed >= r.heartbeatTimeout {
		r.heartbeatElapsed++
		r.Step(pb.Message{From: r.id, Type: pb.MsgBeat})
	}
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
	r.step = stepCandidate
	r.reset(r.Term + 1)
	r.tick = r.tickElection
	r.Vote = r.id //候选人会默认投票给自己
	r.state = StateCandidate
	r.logger.Infof("%x became candidate at term %d", r.id, r.Term)
}

//初始化成为预候选人信息
func (r *raft) becomePreCandidate() {
	// TODO(xiangli) remove the panic when the raft implementation is stable
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
	r.step = stepLeader
	r.reset(r.Term)
	r.tick = r.tickHeartbeat //发送心跳消息
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
	return r.prs.TallyVotes() //计算投票结果
}

//执行竞选的状态
func (r *raft) Step(m pb.Message) error {
	// Handle the message term, which may result in our stepping down to a follower.
	switch {
	case m.Term > r.Term: //发送端的任期号要大于本地端
		switch {
		default:
			//说明是领导者发送心跳的消息了,避免发送竞选,成为候选人
			if m.Type == pb.MsgHeartbeat {
				//设置任期号和领导者
				r.becomeFollower(m.Term, m.From)
			} else {
				//这时候还没选出领导者,设置任期号
				r.becomeFollower(m.Term, None)
			}
		}
	}

	switch m.Type {
	case pb.MsgHup: //新一轮选举的时候竞选,晋升成为候选人
		if r.preVote {
			r.hup(campaignPreElection)
		} else {
			r.hup(campaignElection)
		}

	case pb.MsgVote, pb.MsgPreVote:
		// We can vote if this is a repeat of a vote we've already cast...
		canVote := r.Vote == m.From || //这个条件也说明了成为了候选人之后是无法投票的，因为把票数投给了自己
			(r.Vote == None && r.lead == None) || //一般为follower状态并且没有领导者可以发起投票
			// ...or this is a PreVote for a future term...
			(m.Type == pb.MsgPreVote && m.Term > r.Term)
		if canVote {
			// When responding to Msg{Pre,}Vote messages we include the term
			// from the message, not the local term. To see why, consider the
			// case where a single node was previously partitioned away and
			// it's local term is now out of date. If we include the local term
			// (recall that for pre-votes we don't update the local term), the
			// (pre-)campaigning node on the other end will proceed to ignore
			// the message (it ignores all out of date messages).
			// The term in the original message and current local term are the
			// same in the case of regular votes, but different for pre-votes.
			r.send(pb.Message{To: m.From, Term: m.Term, Type: voteRespMsgType(m.Type)})
			if m.Type == pb.MsgVote {
				// Only record real votes.
				r.electionElapsed = 0
				r.Vote = m.From //记录投票信息
			}
		} else {
			//投反对票
			r.send(pb.Message{To: m.From, Term: r.Term, Type: voteRespMsgType(m.Type), Reject: true})
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
	// These message types do not require any progress for m.From.
	//领导者接收自身的消息
	switch m.Type {
	case pb.MsgBeat: //接收到探活消息
		r.bcastHeartbeat()
	case pb.MsgCheckQuorum:
		// The leader should always see itself as active. As a precaution, handle
		// the case in which the leader isn't in the configuration any more (for
		// example if it just removed itself).
		//
		// TODO(tbg): I added a TODO in removeNode, it doesn't seem that the
		// leader steps down when removing itself. I might be missing something.
		//leader自身要设置处于活跃状态
		if pr := r.prs.Progress[r.id]; pr != nil {
			pr.RecentActive = true
		}
		//超过半数处于不活跃状态
		if !r.prs.QuorumActive() {
			r.logger.Warningf("%x stepped down to follower since quorum is not active", r.id)
			r.becomeFollower(r.Term, None) //leader会变为follower
		}
		// Mark everyone (but ourselves) as inactive in preparation for the next
		// CheckQuorum.
		//设置所有节点连接处于不活跃状态，迎接新的领导者
		r.prs.Visit(func(id uint64, pr *tracker.Progress) {
			if id != r.id {
				pr.RecentActive = false
			}
		})
		return nil
	}

	// All other message types require a progress for m.From (pr).
	//领导者接收来自follower节点的消息
	pr := r.prs.Progress[m.From]
	if pr == nil {
		r.logger.Debugf("%x no progress available for %x", r.id, m.From)
		return nil
	}
	switch m.Type {
	case pb.MsgHeartbeatResp: //心跳的答复消息
		fmt.Println("接收到心跳")
	}
	return nil
}

// stepCandidate is shared by StateCandidate and StatePreCandidate; the difference is
// whether they respond to MsgVoteResp or MsgPreVoteResp.
func stepCandidate(r *raft, m pb.Message) error {
	// Only handle vote responses corresponding to our candidacy (while in
	// StateCandidate, we may get stale MsgPreVoteResp messages in this term from
	// our pre-candidate state).
	var myVoteRespType pb.MessageType
	if r.state == StatePreCandidate {
		myVoteRespType = pb.MsgPreVoteResp
	} else {
		myVoteRespType = pb.MsgVoteResp
	}
	switch m.Type {
	case myVoteRespType:
		gr, rj, res := r.poll(m.From, m.Type, !m.Reject)
		r.logger.Infof("%x has received %d %s votes and %d vote rejections", r.id, gr, m.Type, rj)
		switch res {
		case quorum.VoteWon:
			if r.state == StatePreCandidate {
				r.campaign(campaignElection)
			} else {
				r.becomeLeader()
			}
		case quorum.VoteLost:
			//竞选失败，成为跟随者,如果没有产生领导者则会发起新一轮选举
			// pb.MsgPreVoteResp contains future term of pre-candidate
			// m.Term > r.Term; reuse r.Term
			r.becomeFollower(r.Term, None)
		}
	}
	return nil
}

func stepFollower(r *raft, m pb.Message) error {
	switch m.Type {
	case pb.MsgHeartbeat:
		//如果follower长时间没收到心跳消息，
		//就会假定Leader已经不存在或者发生了故障，于是会发起一次新的选举
		r.electionElapsed = 0
		r.lead = m.From
		r.handleHeartbeat(m)
	}
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

// pastElectionTimeout returns true iff r.electionElapsed is greater
// than or equal to the randomized election timeout in
// [electiontimeout, 2 * electiontimeout - 1].
//判断选举是否超时，超时会触发新一轮选举
func (r *raft) pastElectionTimeout() bool {
	return r.electionElapsed >= r.randomizedElectionTimeout
}

//生成一个随机的选举超时数
func (r *raft) resetRandomizedElectionTimeout() {
	r.randomizedElectionTimeout = r.electionTimeout + globalRand.Intn(r.electionTimeout)
}

//响应心跳
func (r *raft) handleHeartbeat(m pb.Message) {
	r.send(pb.Message{To: m.From, Type: pb.MsgHeartbeatResp, Context: m.Context})
}

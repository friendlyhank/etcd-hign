package tracker

import "github.com/friendlyhank/etcd-hign/raftmodule/raft/quorum"

//raft统计票数相关

// Config reflects the configuration tracked in a ProgressTracker.
type Config struct {
	Voters quorum.JointConfig //这里为啥要搞两个去投票不太理解
}

// Clone returns a copy of the Config that shares no memory with the original.
//复制配置信息
func (c *Config) Clone() Config {
	clone := func(m map[uint64]struct{}) map[uint64]struct{} {
		if m == nil {
			return nil
		}
		mm := make(map[uint64]struct{}, len(m))
		for k := range m {
			mm[k] = struct{}{}
		}
		return mm
	}
	return Config{
		Voters: quorum.JointConfig{clone(c.Voters[0]), clone(c.Voters[1])},
	}
}

// ProgressTracker tracks the currently active configuration and the information
// known about the nodes and learners in it. In particular, it tracks the match
// index for each peer which in turn allows reasoning about the committed index.
//程序投票统计
type ProgressTracker struct {
	Config
	Progress ProgressMap

	Votes map[uint64]bool
}

// MakeProgressTracker initializes a ProgressTracker.
func MakeProgressTracker(maxInflight int) ProgressTracker {
	p := ProgressTracker{
		Config: Config{
			Voters: quorum.JointConfig{
				quorum.MajorityConfig{},
				nil, // only populated when used
			},
		},
		Votes:    map[uint64]bool{},
		Progress: map[uint64]*Progress{},
	}
	return p
}

func insertionSort(sl []uint64) {
	a, b := 0, len(sl)
	for i := a + 1; i < b; i++ {
		for j := i; j > a && sl[j] < sl[j-1]; j-- {
			sl[j], sl[j-1] = sl[j-1], sl[j]
		}
	}
}

// Visit invokes the supplied closure for all tracked progresses in stable order.
func (p *ProgressTracker) Visit(f func(id uint64, pr *Progress)) {
	n := len(p.Progress)
	// We need to sort the IDs and don't want to allocate since this is hot code.
	// The optimization here mirrors that in `(MajorityConfig).CommittedIndex`,
	// see there for details.
	//这是一个热门方法调用，经常用于广播消息，所以做了优化用数组,当数量大于7的时候才用切片
	var sl [7]uint64
	ids := sl[:]
	if len(sl) >= n {
		ids = sl[:n]
	} else {
		ids = make([]uint64, n)
	}
	for id := range p.Progress {
		n--
		ids[n] = id
	}
	insertionSort(ids)
	for _, id := range ids {
		f(id, p.Progress[id])
	}
}

// RecordVote records that the node with the given id voted for this Raft
// instance if v == true (and declined it otherwise).
//记录投票
func (p *ProgressTracker) RecordVote(id uint64, v bool) {
	_, ok := p.Votes[id]
	if !ok {
		p.Votes[id] = v
	}
}

// TallyVotes returns the number of granted and rejected Votes, and whether the
// election outcome is known.
//计算投票的结果
func (p *ProgressTracker) TallyVotes() (granted int, rejected int, _ quorum.VoteResult) {
	// Make sure to populate granted/rejected correctly even if the Votes slice
	// contains members no longer part of the configuration. This doesn't really
	// matter in the way the numbers are used (they're informational), but might
	// as well get it right.
	for id, _ := range p.Progress {
		v, voted := p.Votes[id]
		if !voted {
			continue
		}
		if v {
			granted++ //赞成的票数
		} else {
			rejected++ //反对的票数
		}
	}
	result := p.Voters.VoteResult(p.Votes)
	return granted, rejected, result
}

package confchange

import (
	pb "github.com/friendlyhank/etcd-hign/raftmodule/raft/raftpb"
	"github.com/friendlyhank/etcd-hign/raftmodule/raft/tracker"
)

//raft修改配置信息

// Changer facilitates configuration changes. It exposes methods to handle
// simple and joint consensus while performing the proper validation that allows
// refusing invalid configuration changes before they affect the active
// configuration.
type Changer struct {
	Tracker tracker.ProgressTracker
}

// Simple carries out a series of configuration changes that (in aggregate)
// mutates the incoming majority config Voters[0] by at most one. This method
// will return an error if that is not the case, if the resulting quorum is
// zero, or if the configuration is in a joint state (i.e. if there is an
// outgoing configuration).
//简单的配置更改
func (c Changer) Simple(ccs ...pb.ConfChangeSingle) (tracker.Config, tracker.ProgressMap, error) {
	//校验配置并且复制旧的配置
	cfg, prs, err := c.checkAndCopy()
	if err != nil {

	}

	//申请修改配置设置
	if err := c.apply(nil, prs, ccs...); err != nil {
	}
	return cfg, prs, nil
}

// apply a change to the configuration. By convention, changes to voters are
// always made to the incoming majority config Voters[0]. Voters[1] is either
// empty or preserves the outgoing majority configuration while in a joint state.
//申请修改配置设置
func (c Changer) apply(cfg *tracker.Config, prs tracker.ProgressMap, ccs ...pb.ConfChangeSingle) error {
	for _, cc := range ccs {
		switch cc.Type {
		case pb.ConfChangeAddNode:
			c.makeVoter(cfg, prs, cc.NodeID)
		}
	}
	return nil
}

// makeVoter adds or promotes the given ID to be a voter in the incoming
// majority config.
//设置投票者
func (c Changer) makeVoter(cfg *tracker.Config, prs tracker.ProgressMap, id uint64) {
	pr := prs[id]
	if pr == nil {
		c.initProgress(cfg, prs, id, false)
	}
}

// initProgress initializes a new progress for the given node or learner.
func (c Changer) initProgress(cfg *tracker.Config, prs tracker.ProgressMap, id uint64, isLearner bool) {
	prs[id] = &tracker.Progress{}
}

// checkAndCopy copies the tracker's config and progress map (deeply enough for
// the purposes of the Changer) and returns those copies. It returns an error
// if checkInvariants does.
func (c Changer) checkAndCopy() (tracker.Config, tracker.ProgressMap, error) {
	prs := tracker.ProgressMap{}
	//把配置信息先复制一份新的
	for id, pr := range c.Tracker.Progress {
		// A shallow copy is enough because we only mutate the Learner field.
		ppr := *pr
		prs[id] = &ppr
	}
	return tracker.Config{}, prs, nil
}

package raft

type RawNode struct {
	raft *raft
}

func NewRawNode(config *Config) (*RawNode, error) {
	r := newRaft(config)
	rn := &RawNode{
		raft: r,
	}
	return rn, nil
}

// Tick advances the internal logical clock by a single tick.
func (rn *RawNode) Tick() {
	rn.raft.tick()
}

func (rn *RawNode) readyWithoutAccept() Ready {
	return newReady(rn.raft)
}

//是否处在ready状态可以发送消息
func (rn *RawNode) HasReady() bool {
	r := rn.raft
	if len(r.msgs) > 0 {
		return true
	}
	return false
}

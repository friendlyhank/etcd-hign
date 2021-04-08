package raft

type RawNode struct {
	raft *raft
}

func NewRawNode() (*RawNode, error) {
	r := newRaft()
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

func (rn *RawNode) HasReady() bool {
	return false
}

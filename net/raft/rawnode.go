package raft

type RawNode struct {
	raft *raft
}

func (rn *RawNode) readyWithoutAccept() Ready {
	return newReady(rn.raft)
}

func (rn *RawNode) HasReady() bool {
	return false
}

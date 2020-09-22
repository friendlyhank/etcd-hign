package raft

type RawNode struct {
}

func (rn *RawNode) readyWithoutAccept() Ready {
	return newReady()
}

func (rn *RawNode) HasReady() bool {
	return false
}

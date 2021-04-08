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

func (rn *RawNode) readyWithoutAccept() Ready {
	return newReady(rn.raft)
}

func (rn *RawNode) HasReady() bool {
	return false
}

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

// acceptReady is called when the consumer of the RawNode has decided to go
// ahead and handle a Ready. Nothing must alter the state of the RawNode between
// this call and the prior call to Ready().
//消息已经提交到ready通道,把消息重置掉
func (rn *RawNode) acceptReady(rd Ready) {
	rn.raft.msgs = nil
}

//是否处在ready状态可以发送消息
func (rn *RawNode) HasReady() bool {
	r := rn.raft
	if len(r.msgs) > 0 {
		return true
	}
	return false
}

package raft

type Ready struct {
}

type Node interface {
	Ready() <-chan Ready
}

type node struct {
	readyc chan Ready

	rn *RawNode
}

func StartNode() Node {
	n := newNode()

	//启动node
	//Ready在这里
	go n.run()
	return &n
}

func newNode() node {
	return node{
		readyc: make(chan Ready),
	}
}

func (n *node) run() {
	var readyc chan Ready
	var rd Ready
	for {
		rd = n.rn.readyWithoutAccept()
		readyc = n.readyc

		select {
		case readyc <- rd:
		}
	}
}

// newReady- 在这里去new Ready
func newReady() Ready {
	rd := Ready{}
	return rd
}

func (n *node) Ready() <-chan Ready { return n.readyc }
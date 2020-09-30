package rafthttp

import "time"

const (
	ConnReadTimeout  = 5 * time.Second
	ConnWriteTimeout = 5 * time.Second
)

type Peer interface {
	// attachOutgoingConn attaches the outgoing connection to the peer for
	// stream usage. After the call, the ownership of the outgoing
	// connection hands over to the peer. The peer will close the connection
	// when it is no longer used.
	//这是节点stream的连接
	attachOutgoingConn(conn *outgoingConn)
}

type peer struct{}

func (p *peer) attachOutgoingConn(conn *outgoingConn) {

}

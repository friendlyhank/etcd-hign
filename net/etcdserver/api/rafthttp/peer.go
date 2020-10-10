package rafthttp

import (
	"time"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
)

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

type peer struct {
	lg *zap.Logger

	localID types.ID //本地节点唯一id
	// id of the remote raft peer node 除本地节点外的某个节点
	id types.ID

	msgAppV2Writer *streamWriter
	writer         *streamWriter
}

func startPeer(t *Transport, urls types.URLs, peerID types.ID) *peer {
	p := &peer{
		localID:        t.ID,
		id:             peerID,
		msgAppV2Writer: startStreamWriter(t.Logger, t.ID, peerID),
		writer:         startStreamWriter(t.Logger, t.ID, peerID),
	}
	return p
}

func (p *peer) attachOutgoingConn(conn *outgoingConn) {

}

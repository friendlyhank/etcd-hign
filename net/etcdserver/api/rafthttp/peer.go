package rafthttp

import (
	"fmt"
	"time"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
)

const (
	ConnReadTimeout  = 5 * time.Second
	ConnWriteTimeout = 5 * time.Second

	streamAppV2 = "streamMsgAppV2"
	streamMsg   = "streamMsg"
	pipelineMsg = "pipeline"
)

type Peer interface {
	// send sends the message to the remote peer. The function is non-blocking
	// and has no promise that the message will be received by the remote.
	// When it fails to send message out, it will report the status to underlying
	// raft.
	send(m raftpb.Message)
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
	msgAppV2Reader *streamReader
	msgAppReader   *streamReader
}

func startPeer(t *Transport, urls types.URLs, peerID types.ID) *peer {
	p := &peer{
		localID:        t.ID,
		id:             peerID,
		msgAppV2Writer: startStreamWriter(t.Logger, t.ID, peerID), //启动streamv2写入流
		writer:         startStreamWriter(t.Logger, t.ID, peerID), //启动stream写入流
	}

	p.msgAppV2Reader = &streamReader{
		lg:     t.Logger,
		peerID: peerID,
		typ:    streamTypeMsgAppV2,
		tr:     t,
	}

	p.msgAppReader = &streamReader{
		lg:     t.Logger,
		peerID: peerID,
		typ:    streamTypeMessage,
		tr:     t,
	}

	p.msgAppV2Reader.start()
	p.msgAppReader.start()

	return p
}

func (p *peer) send(m raftpb.Message) {
	writec, name := p.pick(m)
	select {
	case writec <- m:
		fmt.Println(name)
	default:

	}
}

func (p *peer) attachOutgoingConn(conn *outgoingConn) {
	var ok bool
	switch conn.t {
	case streamTypeMsgAppV2:
		ok = p.msgAppV2Writer.attach(conn)
	case streamTypeMessage:
		ok = p.writer.attach(conn)
	default:
	}
	if !ok {

	}
}

func (p *peer) pick(m raftpb.Message) (writec chan<- raftpb.Message, picked string) {
	return p.writer.msgc, streamMsg
}

package rafthttp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
)

const (
	// ConnReadTimeout and ConnWriteTimeout are the i/o timeout set on each connection rafthttp pkg creates.
	// A 5 seconds timeout is good enough for recycling bad connections. Or we have to wait for
	// tcp keepalive failing to detect a bad connection, which is at minutes level.
	// For long term streaming connections, rafthttp pkg sends application level linkHeartbeatMessage
	// to keep the connection alive.
	// For short term pipeline connections, the connection MUST be killed to avoid it being
	// put back to http pkg connection pool.
	ConnReadTimeout  = 5 * time.Second
	ConnWriteTimeout = 5 * time.Second

	recvBufSize = 4096

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

	status *peerStatus

	picker *urlPicker

	msgAppV2Writer *streamWriter
	writer         *streamWriter
	pipeline       *pipeline
	msgAppV2Reader *streamReader
	msgAppReader   *streamReader

	recvc chan raftpb.Message
	propc chan raftpb.Message

	mu sync.Mutex

	cancel context.CancelFunc // cancel pending works in go routine created by peer.
	stopc  chan struct{}
}

func startPeer(t *Transport, urls types.URLs, peerID types.ID) *peer {
	if t.Logger != nil {
		t.Logger.Info("starting remote peer", zap.String("remote-peer-id", peerID.String()))
	}
	defer func() {
		if t.Logger != nil {
			t.Logger.Info("started remote peer", zap.String("remote-peer-id", peerID.String()))
		}
	}()

	status := newPeerStatus(t.Logger, t.ID, peerID)
	picker := newURLPicker(urls)
	errorc := t.ErrorC

	pipeline := &pipeline{
		peerID: peerID,
		tr:     t,
		picker: picker,
		status: status,
		errorc: errorc,
	}
	//启动pipeline start
	pipeline.start()

	p := &peer{
		lg:             t.Logger,
		localID:        t.ID,
		id:             peerID,
		status:         status,
		picker:         picker,
		msgAppV2Writer: startStreamWriter(t.Logger, t.ID, peerID), //启动streamv2写入流
		writer:         startStreamWriter(t.Logger, t.ID, peerID), //启动stream写入流
		pipeline:       pipeline,
		recvc:          make(chan raftpb.Message, recvBufSize),
		stopc:          make(chan struct{}),
	}

	//ctx, cancel := context.WithCancel(context.Background())
	//p.cancel = cancel
	go func() {
		for {
			select {
			case mm := <-p.recvc:
				fmt.Println(mm)
			case <-p.stopc:
				return
			}
		}
	}()

	// r.Process might block for processing proposal when there is no leader.
	// Thus propc must be put into a separate routine with recvc to avoid blocking
	// processing other raft messages.
	go func() {
		for {
			select {
			case mm := <-p.propc:
				fmt.Println(mm)
			case <-p.stopc:
				return
			}
		}
	}()

	p.msgAppV2Reader = &streamReader{
		lg:     t.Logger,
		peerID: peerID,
		typ:    streamTypeMsgAppV2,
		tr:     t,
		picker: picker,
		status: status,
		recvc:  p.recvc,
		propc:  p.propc,
		rl:     rate.NewLimiter(t.DialRetryFrequency, 1), //拨号限制频率
	}

	p.msgAppReader = &streamReader{
		lg:     t.Logger,
		peerID: peerID,
		typ:    streamTypeMessage,
		tr:     t,
		picker: picker,
		status: status,
		recvc:  p.recvc,
		propc:  p.propc,
		rl:     rate.NewLimiter(t.DialRetryFrequency, 1), //拨号限制频率
	}

	p.msgAppV2Reader.start()
	p.msgAppReader.start()

	return p
}

func (p *peer) send(m raftpb.Message) {
	//writec, name := p.pick(m)
	//select {
	//case writec <- m:
	//	fmt.Println(name)
	//default:
	//
	//}
}

func (p *peer) attachOutgoingConn(conn *outgoingConn) {
	var ok bool
	switch conn.t {
	case streamTypeMsgAppV2:
		ok = p.msgAppV2Writer.attach(conn)
	case streamTypeMessage:
		ok = p.writer.attach(conn)
	default:
		if p.lg != nil {
			p.lg.Panic("unknown stream type", zap.String("type", conn.t.String()))
		}
	}
	if !ok {
		conn.Close()
	}
}

func (p *peer) pick(m raftpb.Message) (writec chan<- raftpb.Message, picked string) {
	return p.writer.msgc, streamMsg
}

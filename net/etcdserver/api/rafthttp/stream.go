package rafthttp

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"go.uber.org/zap"
)

const (
	streamTypeMessage  streamType = "message"
	streamTypeMsgAppV2 streamType = "msgappv2"
)

type streamType string //流的类型 v2|v3的流

func (t streamType) endpoint(lg *zap.Logger) string {
	switch t {
	case streamTypeMsgAppV2:
		return path.Join(RaftStreamPrefix, "msgapp")
	case streamTypeMessage:
		return path.Join(RaftStreamPrefix, "message")
	default:
		if lg != nil {
			lg.Panic("unhandled stream type", zap.String("stream-type", t.String()))
		}
		return ""
	}
}

func (t streamType) String() string {
	switch t {
	case streamTypeMsgAppV2:
		return "stream MsgApp v2"
	case streamTypeMessage:
		return "stream Message"
	default:
		return "unknown stream"
	}
}

type outgoingConn struct {
	t streamType //流的类型
	io.Writer
	http.Flusher
	localID types.ID
	peerID  types.ID
}

type streamWriter struct {
	lg *zap.Logger

	localID types.ID
	peerID  types.ID

	mu      sync.Mutex
	working bool //流是否处于工作状态

	msgc  chan raftpb.Message //接收消息
	connc chan *outgoingConn  //获取连接
	stopc chan struct{}
	done  chan struct{}
}

func startStreamWriter(lg *zap.Logger, local, id types.ID) *streamWriter {
	w := &streamWriter{
		localID: local,
		peerID:  id,
		connc:   make(chan *outgoingConn),
	}
	go w.run()
	return w
}

func (cw *streamWriter) run() {
	var (
		msgc       chan raftpb.Message
		heartbeatc <-chan time.Time
		flusher    http.Flusher
	)
	//设置读取心跳时间
	tickc := time.NewTicker(ConnReadTimeout / 3)
	defer tickc.Stop()
	for {
		select {
		case conn := <-cw.connc:
			flusher = conn.Flusher
			heartbeatc, msgc = tickc.C, cw.msgc
			fmt.Println(flusher)
			fmt.Println(msgc)
			fmt.Println(heartbeatc)
		}
	}
}

//将streamWtiter写入chan
func (cw *streamWriter) attach(conn *outgoingConn) bool {
	select {
	case cw.connc <- conn:
		return true
	}
}

package rafthttp

import (
	"fmt"
	"io"
	"net/http"
	"path"

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
	localID types.ID
	peerID  types.ID
	connc   chan *outgoingConn
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
	for {
		select {
		case conn := <-cw.connc:
			fmt.Println(conn)
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

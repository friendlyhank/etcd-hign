package rafthttp

import (
	"fmt"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"go.uber.org/zap"
)

type outgoingConn struct{}

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

package rafthttp

import (
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"
)

const (
	connPerPipeline = 4
)

type pipeline struct {
	peerID types.ID

	tr     *Transport
	picker *urlPicker

	msgc chan raftpb.Message
	// wait for the handling routines
	wg    sync.WaitGroup
	stopc chan struct{}
}

func (p *pipeline) start() {
	p.wg.Add(connPerPipeline)
	for i := 0; i < connPerPipeline; i++ {

	}
	if p.tr != nil && p.tr.Logger != nil {
		p.tr.Logger.Info(
			"started HTTP pipelining with remote peer",
			zap.String("local-member-id", p.tr.ID.String()),
			zap.String("remote-peer-id", p.peerID.String()),
		)
	}
}

func (p *pipeline) handle() {
	defer p.wg.Done()
	for {
		select {
		case m := <-p.msgc:
			fmt.Println(m)
		case <-p.stopc:
			return
		}
	}
}

package rafthttp

import (
	"context"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/friendlyhank/etcd-hign/net/raft"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
)

func TestSendMessage(t *testing.T) {
	//member 1
	tr := &Transport{
		ID:        types.ID(1),
		ClusterID: types.ID(1),
		Raft:      &fakeRaft{},
	}
	tr.Start()
	//httptest测试包测试
	srv := httptest.NewServer(tr.Handler())
	defer srv.Close()

	//member 2
	recvc := make(chan raftpb.Message, 1)
	p := &fakeRaft{recvc: recvc} //构造一个fackRaft用来测试消息的接收
	tr2 := &Transport{
		ID:        types.ID(2),
		ClusterID: types.ID(1),
		Raft:      p,
	}
	tr2.Start()
	srv2 := httptest.NewServer(tr2.Handler())
	defer srv2.Close()

	tr.AddPeer(types.ID(2), []string{srv2.URL})
	defer tr.Stop()
	tr2.AddPeer(types.ID(1), []string{srv.URL})
	defer tr2.Stop()
	if !waitStreamWorking(tr.Get(types.ID(2)).(*peer)) {
		t.Fatalf("stream from 1 to 2 is not in work as expected")
	}

	data := []byte("some data")
	tests := []raftpb.Message{
		{Type: raftpb.MsgProp, From: 1, To: 2, Entries: []raftpb.Entry{{Data: data}}},
		{Type: raftpb.MsgApp, From: 1, To: 2, Term: 1, Index: 3, LogTerm: 0, Entries: []raftpb.Entry{{Index: 4, Term: 1, Data: data}}, Commit: 3},
		{Type: raftpb.MsgAppResp, From: 1, To: 2, Term: 1, Index: 3},
		{Type: raftpb.MsgVote, From: 1, To: 2, Term: 1, Index: 3, LogTerm: 0},
		{Type: raftpb.MsgVoteResp, From: 1, To: 2, Term: 1},
		{Type: raftpb.MsgSnap, From: 1, To: 2, Term: 1, Snapshot: raftpb.Snapshot{Metadata: raftpb.SnapshotMetadata{Index: 1000, Term: 1}, Data: data}},
		{Type: raftpb.MsgHeartbeat, From: 1, To: 2, Term: 1, Commit: 3},
		{Type: raftpb.MsgHeartbeatResp, From: 1, To: 2, Term: 1},
	}
	//测试去发送数据,然后etcd-raft去接收
	for i, tt := range tests {
		tr.Send([]raftpb.Message{tt})
		msg := <-recvc
		if !reflect.DeepEqual(msg, tt) {
			t.Errorf("#%d: msg = %+v, want %+v", i, msg, tt)
		}
		fmt.Println(msg)
	}
}

func waitStreamWorking(p *peer) bool {
	for i := 0; i < 1000; i++ {
		time.Sleep(time.Millisecond)
		if _, ok := p.msgAppV2Writer.writec(); !ok {
			continue
		}
		if _, ok := p.writer.writec(); !ok {
			continue
		}
		return true
	}
	return false
}

type fakeRaft struct {
	recvc     chan<- raftpb.Message
	err       error
	removedID uint64
}

func (p *fakeRaft) Process(ctx context.Context, m raftpb.Message) error {
	select {
	case p.recvc <- m:
	default:
	}
	return p.err
}

func (p *fakeRaft) IsIDRemoved(id uint64) bool { return id == p.removedID }

func (p *fakeRaft) ReportUnreachable(id uint64) {}

func (p *fakeRaft) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}
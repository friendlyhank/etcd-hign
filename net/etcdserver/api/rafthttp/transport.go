package rafthttp

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"
)

// Transporter -网络接口核心
type Transporter interface {
	Handler() http.Handler
}

type Transport struct {
	ID    types.ID          //local member ID 当前节点的唯一id
	mu    sync.RWMutex      //protect the remote and peer map
	peers map[types.ID]Peer //peers map
}

func (t *Transport) Handler() http.Handler {
	/*

	 */
	streamHandler := newStreamHandler()
	mux := http.NewServeMux()
	mux.Handle(RaftStreamPrefix+"/", streamHandler)
	return mux
}

func (t *Transport) Get(id types.ID) Peer {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.peers[id]
}

func (t *Transport) Send(msgs []raftpb.Message) {
	for _, m := range msgs {
		fmt.Println(m)
	}
}

func (t *Transport) AddRemote(id types.ID, us []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
}

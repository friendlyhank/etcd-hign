package rafthttp

import (
	"net/http"
	"sync"

	"go.uber.org/zap"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"
)

// Transporter -网络接口核心
type Transporter interface {
	// Handler returns the HTTP handler of the transporter.
	// A transporter HTTP handler handles the HTTP requests
	// from remote peers.
	// The handler MUST be used to handle RaftPrefix(/raft)
	// endpoint.
	Handler() http.Handler
	// Send sends out the given messages to the remote peers.
	// Each message has a To field, which is an id that maps
	// to an existing peer in the transport.
	// If the id cannot be found in the transport, the message
	// will be ignored.
	Send(m []raftpb.Message)
}

type Transport struct {
	Logger *zap.Logger

	ID    types.ID          //local member ID 当前节点的唯一id
	mu    sync.RWMutex      //protect the remote and peer map
	peers map[types.ID]Peer //peers map
}

func (t *Transport) Start() error {
	t.peers = make(map[types.ID]Peer)
	return nil
}

func (t *Transport) Handler() http.Handler {
	/*
	 */
	streamHandler := newStreamHandler(t, t, t.ID)
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
		if m.To == 0 {
			continue
		}
		to := types.ID(m.To)

		t.mu.RLock()
		p, pok := t.peers[to]
		t.mu.RUnlock()

		if pok {

			continue
		}
	}
}

func (t *Transport) AddRemote(id types.ID, us []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.peers[id]; ok {
		return
	}
	urls, err := types.NewURLs(us)
	if err != nil {
	}
	t.peers[id] = startPeer(t, urls, id)
}

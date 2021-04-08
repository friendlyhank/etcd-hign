package rafthttp

import (
	"net/url"
	"sync"

	"github.com/friendlyhank/etcd-hign/netmodule/pkg/types"
)

type urlPicker struct {
	mu     sync.Mutex
	urls   types.URLs
	picked int
}

func newURLPicker(urls types.URLs) *urlPicker {
	return &urlPicker{
		urls: urls,
	}
}

func (p *urlPicker) update(urls types.URLs) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.urls = urls
	p.picked = 0
}

func (p *urlPicker) pick() url.URL {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.urls[p.picked]
}

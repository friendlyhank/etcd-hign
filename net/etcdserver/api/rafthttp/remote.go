package rafthttp

import (
	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"go.uber.org/zap"
)

//TODO Hank这个到底为啥需要这个
type remote struct {
	lg       *zap.Logger
	localID  types.ID
	id       types.ID
	pipeline *pipeline
}

func startRemote(tr *Transport, urls types.URLs, id types.ID) *remote {
	picker := newURLPicker(urls)
	pipeline := &pipeline{
		peerID: id,
		tr:     tr,
		picker: picker,
	}
	pipeline.start()

	return &remote{
		lg:       tr.Logger,
		localID:  tr.ID,
		id:       id,
		pipeline: pipeline,
	}
}

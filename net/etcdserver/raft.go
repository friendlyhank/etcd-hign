package etcdserver

import (
	"github.com/friendlyhank/etcd-hign/net/etcdserver/api/rafthttp"
)

type raftNode struct {
	raftNodeConfig //这里隐藏很多重要信息 transport,Node
}

type raftNodeConfig struct {
	transport rafthttp.Transporter
}

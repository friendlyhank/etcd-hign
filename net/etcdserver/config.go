package etcdserver

import "github.com/friendlyhank/etcd-hign/net/pkg/types"

type ServerConfig struct {
	Name                string
	InitialPeerURLsMap  types.URLsMap
	InitialClusterToken string
}

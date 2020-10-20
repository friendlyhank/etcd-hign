package etcdserver

import (
	"github.com/friendlyhank/etcd-hign/net/pkg/transport"
	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"go.uber.org/zap"
)

type ServerConfig struct {
	Name                string
	InitialPeerURLsMap  types.URLsMap
	InitialClusterToken string
	PeerTLSInfo         transport.TLSInfo

	// Logger logs server-side operations.
	// If not nil, it disables "capnslog" and uses the given logger.
	Logger *zap.Logger
}

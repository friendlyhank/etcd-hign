package membership

import (
	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"go.uber.org/zap"
)

type RaftCluster struct {
	members map[types.ID]*Member
}

func NewClusterFromURLsMap(lg *zap.Logger, token string, urlsmap types.URLsMap) (*RaftCluster, error) {
	c := NewCluster()
	for name, urls := range urlsmap {
		m := newMember()
		if _,ok :=c.members[m]
	}
}

func NewCluster() *RaftCluster {
	return &RaftCluster{}
}

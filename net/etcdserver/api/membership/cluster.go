package membership

import (
	"fmt"

	"go.etcd.io/etcd/raft"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"go.uber.org/zap"
)

type RaftCluster struct {
	members map[types.ID]*Member
}

func NewClusterFromURLsMap(lg *zap.Logger, token string, urlsmap types.URLsMap) (*RaftCluster, error) {
	c := NewCluster()
	for name, urls := range urlsmap {
		m := NewMember(name, urls, token, nil)
		if _, ok := c.members[m.ID]; ok {
			return nil, fmt.Errorf("member exists with identical ID %v", m)
		}
		if uint64(m.ID) == raft.None {
			return nil, fmt.Errorf("cannot use %x as member id", raft.None)
		}
		c.members[m.ID] = m
	}
	return c, nil
}

func NewCluster() *RaftCluster {
	return &RaftCluster{}
}

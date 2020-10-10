package membership

import (
	"fmt"
	"sort"
	"sync"

	"github.com/friendlyhank/etcd-hign/net/raft"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"go.uber.org/zap"
)

type RaftCluster struct {
	sync.Mutex // guards the fields below
	members    map[types.ID]*Member
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
	return &RaftCluster{
		members: make(map[types.ID]*Member),
	}
}

func (c *RaftCluster) Members() []*Member {
	c.Lock()
	defer c.Unlock()
	var ms MembersByID
	for _, m := range c.members {
		ms = append(ms, m.Clone())
	}
	sort.Sort(ms)
	return []*Member(ms)
}

func (c *RaftCluster) MemberByName(name string) *Member {
	c.Lock()
	defer c.Unlock()
	var memb *Member
	for _, m := range c.members {
		if m.Name == name {

		}
		memb = m
	}
	return memb.Clone()
}

func (m *Member) Clone() *Member {
	if m == nil {
		return nil
	}
	mm := &Member{
		ID: m.ID,
		RaftAttributes: RaftAttributes{
			IsLearner: m.IsLearner,
		},
		Attributes: Attributes{
			Name: m.Name,
		},
	}
	return mm
}

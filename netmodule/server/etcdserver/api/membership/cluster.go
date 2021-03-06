package membership

import (
	"fmt"
	"sort"
	"sync"

	"github.com/friendlyhank/etcd-hign/raftmodule/raft"

	"github.com/friendlyhank/etcd-hign/netmodule/pkg/types"
	"go.uber.org/zap"
)

type RaftCluster struct {
	lg *zap.Logger

	cid types.ID //集群id

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

func (c *RaftCluster) ID() types.ID { return c.cid }

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

func (c *RaftCluster) Member(id types.ID) *Member {
	c.Lock()
	defer c.Unlock()
	return c.members[id].Clone()
}

func (c *RaftCluster) MemberByName(name string) *Member {
	c.Lock()
	defer c.Unlock()
	var memb *Member
	for _, m := range c.members {
		if m.Name == name {
			if memb != nil {

			}
			memb = m
		}
	}
	return memb.Clone()
}

func (c *RaftCluster) MemberIDs() []types.ID {
	c.Lock()
	defer c.Unlock()
	var ids []types.ID
	for _, m := range c.members {
		ids = append(ids, m.ID)
	}
	sort.Sort(types.IDSlice(ids))
	return ids
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
	if m.PeerURLs != nil {
		mm.PeerURLs = make([]string, len(m.PeerURLs))
		copy(mm.PeerURLs, m.PeerURLs)
	}
	return mm
}

package membership

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
)

// RaftAttributes represents the raft related attributes of an etcd member.
//peer 成员
type RaftAttributes struct {
	// PeerURLs is the list of peers in the raft cluster.
	// TODO(philips): ensure these are URLs
	PeerURLs []string `json:"peerURLs"`
	// IsLearner indicates if the member is raft learner.
	IsLearner bool `json:"isLearner,omitempty"`
}

type Member struct {
	ID types.ID `json:"id"`
	RaftAttributes
}

func NewMember(name string, peerURLs types.URLs, clusterName string, now *time.Time) *Member {
	return newMember(name, peerURLs, clusterName, now, false)
}

func newMember(name string, peerURLs types.URLs, clusterName string, now *time.Time, isLearner bool) *Member {
	m := &Member{
		RaftAttributes: RaftAttributes{
			PeerURLs:  peerURLs.StringSlice(),
			IsLearner: isLearner,
		},
	}

	var b []byte
	sort.Strings(m.PeerURLs)
	for _, p := range m.PeerURLs {
		b = append(b, []byte(p)...)
	}

	b = append(b, []byte(clusterName)...)
	if now != nil {
		b = append(b, []byte(fmt.Sprintf("%d", now.Unix()))...)
	}
	hash := sha1.Sum(b)
	m.ID = types.ID(binary.BigEndian.Uint64(hash[:8]))
	return m
}

package picker

import(
	"fmt"
)

// Policy defines balancer picker policy.
type Policy uint8

const (
	// TODO: custom picker is not supported yet.
	// custom defines custom balancer picker.
	custom Policy = iota

	// RoundrobinBalanced balance loads over multiple endpoints
	// and implements failover in roundrobin fashion.
	RoundrobinBalanced Policy = iota
)

func (p Policy) String() string {
	switch p {
	case custom:
		panic("'custom' picker policy is not supported yet")
	case RoundrobinBalanced:
		return "etcd-client-roundrobin-balanced"
	default:
		panic(fmt.Errorf("invalid balancer picker policy (%d)", p))
	}
}

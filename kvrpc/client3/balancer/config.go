package balancer

import "github.com/friendlyhank/etcd-hign/kvrpc/client3/balancer/picker"

type Config struct{
	// Policy configures balancer policy.
	Policy picker.Policy

	// Name defines an additional name for balancer.
	// Useful for balancer testing to avoid register conflicts.
	// If empty, defaults to policy name.
	Name string
}

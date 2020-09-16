package etcdmain

import "github.com/friendlyhank/etcd-hign/net/embed"

type config struct{
	ec embed.Config
}

func newConfig() *config{
	cfg :=&config{
		ec:*embed.NewConfig(),
	}
	return cfg
}

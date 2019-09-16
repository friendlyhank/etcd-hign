package balancer

import (
	"google.golang.org/grpc/balancer"
)

type builder struct{
	cfg Config
}

//RegisterBuilder -builder
func RegisterBuilder(cfg Config) {
	bb :=&builder{cfg}
	balancer.Register(bb)
}

func (b *builder)Build(cc balancer.ClientConn, opts balancer.BuildOptions) balancer.Balancer{
	return nil
}

func (b *builder)Name() string{
	return nil
}

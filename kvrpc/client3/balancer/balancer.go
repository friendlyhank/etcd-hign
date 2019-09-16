package balancer

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
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
	bb := &baseBalancer{}
	return bb
}

func (b *builder)Name() string{return b.cfg.Name}


type baseBalancer struct{

}

// HandleResolvedAddrs implements "grpc/balancer.Balancer" interface.
// gRPC sends initial or updated resolved addresses from "Build".
//balancer 地址句柄 HandleResolvedAddrs实现接口grpc/balancer.Balancer
func (bb *baseBalancer) HandleResolvedAddrs(addrs []resolver.Address, err error) {

}

// HandleSubConnStateChange implements "grpc/balancer.Balancer" interface.
//balancer 连接状态句柄 HandleSubConnStateChange实现接口grpc/balancer.Balancer
func (bb *baseBalancer) HandleSubConnStateChange(sc balancer.SubConn, s connectivity.State) {

}

// Close implements "grpc/balancer.Balancer" interface.
// Close is a nop because base balancer doesn't have internal state to clean up,
// and it doesn't need to call RemoveSubConn for the SubConns.
//Close实现接口grpc/balancer.Balancer
func (bb *baseBalancer) Close() {
	// TODO
}
package balancer

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
	"hank.com/etcd-3.3.12-hign/kvrpc/client3/balancer/picker"
	"strconv"
	"sync"
	"time"
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
	bb := &baseBalancer{
		id:strconv.FormatInt(time.Now().UnixNano(),36),
		policy:b.cfg.Policy,
		name:b.cfg.Name,

		addrToSc: make(map[resolver.Address]balancer.SubConn),
		scToAddr: make(map[balancer.SubConn]resolver.Address),
		scToSt:   make(map[balancer.SubConn]connectivity.State),

		currentConn:nil,
	}
	bb.mu.Lock()
	bb.currentConn = cc
	bb.mu.Unlock()
	return bb
}

func (b *builder)Name() string{return b.cfg.Name}


type baseBalancer struct{
	id string
	policy picker.Policy
	name string

	mu sync.Mutex

	addrToSc map[resolver.Address]balancer.SubConn
	scToAddr map[balancer.SubConn]resolver.Address
	scToSt   map[balancer.SubConn]connectivity.State

	currentConn  balancer.ClientConn
	currentState connectivity.State //连接状态
	csEvltr      *connectivityStateEvaluator
}

// HandleResolvedAddrs implements "grpc/balancer.Balancer" interface.
// gRPC sends initial or updated resolved addresses from "Build".
//balancer 地址句柄 HandleResolvedAddrs实现接口grpc/balancer.Balancer
func (bb *baseBalancer) HandleResolvedAddrs(addrs []resolver.Address, err error) {
	if err != nil{
		return
	}

	bb.mu.Lock()
	defer bb.mu.Unlock()

	resolved := make(map[resolver.Address]struct{})
	for _,addr := range addrs{
		resolved[addr] = struct{}{}
		if _,ok := bb.addrToSc[addr];!ok{
			sc,err := bb.currentConn.NewSubConn([]resolver.Address{addr},balancer.NewSubConnOptions{})
			if err != nil{
				continue
			}
			bb.addrToSc[addr]=sc
			bb.scToAddr[sc]=addr
			bb.scToSt[sc] = connectivity.Idle
			sc.Connect()
		}
	}

	for addr, sc := range bb.addrToSc{
		if _,ok := resolved[addr];!ok{
			// was removed by resolver or failed to create subconn
			//resolver被移除或者创建subconn失败
			bb.currentConn.RemoveSubConn(sc)
			delete(bb.addrToSc,addr)
		}
	}
}

// HandleSubConnStateChange implements "grpc/balancer.Balancer" interface.
//balancer 连接状态句柄 HandleSubConnStateChange实现接口grpc/balancer.Balancer
func (bb *baseBalancer) HandleSubConnStateChange(sc balancer.SubConn, s connectivity.State) {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	old,ok := bb.scToSt[sc]
	if !ok{
		return
	}

	bb.scToSt[sc] =s
	switch s {
	case connectivity.Idle:
		sc.Connect()
	case connectivity.Shutdown:
		delete(bb.scToAddr, sc)
		delete(bb.scToSt, sc)
	}

	//oldAggrState := bb.currentState
	bb.currentState = bb.csEvltr.recordTransition(old, s)

}

// Close implements "grpc/balancer.Balancer" interface.
// Close is a nop because base balancer doesn't have internal state to clean up,
// and it doesn't need to call RemoveSubConn for the SubConns.
//Close实现接口grpc/balancer.Balancer
func (bb *baseBalancer) Close() {
	// TODO
}
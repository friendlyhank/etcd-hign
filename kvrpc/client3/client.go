package client3

import(
	"google.golang.org/grpc"
	"context"
	"hank.com/etcd-3.3.12-hign/kvrpc/client3/balancer/picker"
	"hank.com/etcd-3.3.12-hign/kvrpc/client3/balancer/resolver/endpoint"
	"hank.com/etcd-3.3.12-hign/kvrpc/client3/balancer"
	"sync"
	"fmt"
	"github.com/google/uuid"
)

var(
	roundRobinBalancerName = fmt.Sprintf("etcd-%s", picker.RoundrobinBalanced.String())
)

func init(){
	//grpc balancer register
	balancer.RegisterBuilder(balancer.Config{
		Policy: picker.RoundrobinBalanced,
		Name:   roundRobinBalancerName,
	})
}

type Client struct{
	KV

	conn *grpc.ClientConn //grpc链接

	resolverGroup *endpoint.ResolverGroup //grpc resolver build用来设置endpoint
	mu *sync.Mutex

	ctx context.Context
	cancel context.CancelFunc

	callOpts []grpc.CallOption
}

func New(cfg Config)(*Client,error){
	return newClient(&cfg)
}

//根据Config去new Client
func newClient(cfg *Config)(*Client,error){
	if cfg == nil{
		cfg =&Config{}
	}

	client := &Client{
		callOpts:defaultCallOpts,
	}

	// Prepare a 'endpoint://<unique-client-id>/' resolver for the client and create a endpoint target to pass
	// to dial so the client knows to use this resolver.
	//设置resolverGroup 构造resolver用于后面grcp.dial grcp.WithBalancerName
	client.resolverGroup, err = endpoint.NewResolverGroup(fmt.Sprintf("client-%s", uuid.New().String()))
	if err != nil {
		client.cancel()
		return nil, err
	}
	client.resolverGroup.SetEndpoints(cfg.Endpoints)

	if len(cfg.Endpoints) < 1 {
		return nil, fmt.Errorf("at least one Endpoint must is required in client config")
	}

	client.KV = NewKV(client)

	return client,nil
}

func toErr(ctx context.Context, err error) error {
	if err == nil{
		return nil
	}

	return err
}

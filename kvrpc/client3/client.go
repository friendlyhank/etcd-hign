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

	var err error
	client := &Client{
		callOpts:defaultCallOpts,
	}

	if cfg.MaxCallRecvMsgSize > 0 || cfg.MaxCallSendMsgSize > 0{
		if cfg.MaxCallRecvMsgSize > 0 && cfg.MaxCallSendMsgSize > cfg.MaxCallRecvMsgSize {
			return nil, fmt.Errorf("gRPC message recv limit (%d bytes) must be greater than send limit (%d bytes)", cfg.MaxCallRecvMsgSize, cfg.MaxCallSendMsgSize)
		}
		callOpts := []grpc.CallOption{
			defaultFailFast,
			defaultMaxCallSendMsgSize,
			defaultMaxCallRecvMsgSize,
		}
		if cfg.MaxCallSendMsgSize > 0 {
			callOpts[1] = grpc.MaxCallSendMsgSize(cfg.MaxCallSendMsgSize)
		}
		if cfg.MaxCallRecvMsgSize > 0 {
			callOpts[2] = grpc.MaxCallRecvMsgSize(cfg.MaxCallRecvMsgSize)
		}
		client.callOpts = callOpts
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

	dialEndpoint := cfg.Endpoints[0]

	// Use a provided endpoint target so that for https:// without any tls config given, then
	// grpc will assume the certificate server name is the endpoint host.
	//获得grpc conn
	conn, err := client.dialWithBalancer(dialEndpoint, grpc.WithBalancerName(roundRobinBalancerName))

	client.KV = NewKV(client)

	return client,nil
}



func toErr(ctx context.Context, err error) error {
	if err == nil{
		return nil
	}

	return err
}

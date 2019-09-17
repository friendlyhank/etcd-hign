package client3

import(
	"google.golang.org/grpc"
	"context"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"hank.com/etcd-3.3.12-hign/kvrpc/client3/balancer/picker"
	"hank.com/etcd-3.3.12-hign/kvrpc/client3/balancer/resolver/endpoint"
	"hank.com/etcd-3.3.12-hign/kvrpc/client3/balancer"
	"sync"
	"fmt"
	"github.com/google/uuid"
	"crypto/tls"
	"time"
	"net"
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

	cfg           Config //配置文件
	creds         *credentials.TransportCredentials//grpc creds证书相关
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

	client.conn = conn

	client.KV = NewKV(client)

	return client,nil
}

// dialWithBalancer dials the client's current load balanced resolver group.  The scheme of the host
// of the provided endpoint determines the scheme used for all endpoints of the client connection.
func (c *Client) dialWithBalancer(ep string, dopts ...grpc.DialOption) (*grpc.ClientConn, error) {
	_, host, _ := endpoint.ParseEndpoint(ep)
	target := c.resolverGroup.Target(host)
	creds := c.dialWithBalancerCreds(ep)
	return c.dial(target, creds, dopts...)
}

// dial configures and dials any grpc balancer target.
//客户端grpc dial拨号
func (c *Client) dial(target string, creds *credentials.TransportCredentials, dopts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts,err := c.dialSetupOpts(creds,dopts...)
	if err != nil{
		return nil, fmt.Errorf("failed to configure dialer: %v", err)
	}

	dctx := c.ctx

	//grpc dial拨号
	conn,err  := grpc.DialContext(dctx,target,opts...)
	if err != nil{
		return nil,err
	}
	return conn,nil
}

func (c *Client) dialWithBalancerCreds(ep string) *credentials.TransportCredentials {
	_, _, scheme := endpoint.ParseEndpoint(ep)
	creds := c.creds
	if len(scheme) != 0 {
		creds = c.processCreds(scheme)
	}
	return creds
}

//processCreds -处理Creds
func (c *Client) processCreds(scheme string) (creds *credentials.TransportCredentials) {
	creds = c.creds
	switch scheme {
	case "unix":
	case "http":
		creds = nil
	case "https", "unixs":
		if creds != nil {
			break
		}
		tlsconfig := &tls.Config{}
		emptyCreds := credentials.NewTLS(tlsconfig)
		creds = &emptyCreds
	default:
		creds = nil
	}
	return creds
}

// dialSetupOpts gives the dial opts prior to any authentication.
func (c *Client) dialSetupOpts(creds *credentials.TransportCredentials, dopts ...grpc.DialOption) (opts []grpc.DialOption, err error) {
	if c.cfg.DialKeepAliveTime > 0{
		params := keepalive.ClientParameters{
			Time:                c.cfg.DialKeepAliveTime,
			Timeout:             c.cfg.DialKeepAliveTimeout,
			PermitWithoutStream: c.cfg.PermitWithoutStream,
		}
		opts = append(opts, grpc.WithKeepaliveParams(params))
	}
	opts = append(opts, dopts...)

	// Provide a net dialer that supports cancelation and timeout.
	f := func(dialEp string, t time.Duration) (net.Conn, error) {
		proto, host, _ := endpoint.ParseEndpoint(dialEp)
		select {
		case <-c.ctx.Done():
			return nil, c.ctx.Err()
		default:
		}
		dialer := &net.Dialer{Timeout: t}
		return dialer.DialContext(c.ctx, proto, host)
	}
	opts = append(opts, grpc.WithDialer(f))

	if creds != nil{
		opts =append(opts,grpc.WithTransportCredentials(*creds))
	}else{
		opts = append(opts,grpc.WithInsecure())
	}
	return opts,nil
}

func toErr(ctx context.Context, err error) error {
	if err == nil{
		return nil
	}

	return err
}

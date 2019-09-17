package endpoint

import (
	"fmt"
	"google.golang.org/grpc/resolver"
	"net/url"
	"strings"
	"sync"
)

const scheme = "endpoint"

var (
	targetPrefix = fmt.Sprintf("%s://", scheme)

	bldr *builder
)

/**
 *客户端endpoint终端类,用于grpc resolver builder
 *客户端endpoint终端类 为了设置grpc resolver builder interface
 */
func init() {
	bldr = &builder{
		resolverGroups: make(map[string]*ResolverGroup),
	}
	//grpc resolver balancer Register注册
	resolver.Register(bldr)
}

//builder 用map设置ResolverGroup组
type builder struct {
	mu             sync.RWMutex
	resolverGroups map[string]*ResolverGroup
}

func (b *builder) newResolverGroup(id string) (*ResolverGroup, error) {
	b.mu.RLock()
	_, ok := b.resolverGroups[id]
	b.mu.RUnlock()
	if ok {
		return nil, fmt.Errorf("Endpoint already exists for id: %s", id)
	}

	es := &ResolverGroup{id: id}
	b.mu.Lock()
	b.resolverGroups[id] = es
	b.mu.Unlock()
	return es, nil
}

func (b *builder)getResolverGroup(id string)(*ResolverGroup,error){
	b.mu.RLock()
	es,ok := b.resolverGroups[id]
	b.mu.RUnlock()
	if !ok{
		return nil,fmt.Errorf("ResolverGroup not found for id: %s", id)
	}
	return es,nil
}

func (b *builder)Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error){
	id := target.Authority
	es,err :=b.getResolverGroup(id)
	if err != nil{
		return nil, fmt.Errorf("failed to build resolver: %v", err)
	}
	//build resolver还有设置endpoint.Resolver
	r :=&Resolver{
		endpointID:id,
		cc:cc,
	}
	es.addResolver(r)
	return r,nil
}

func (b *builder)Scheme()string{
	return scheme
}

func (b *builder)close(id string){
	b.mu.Lock()
	delete(b.resolverGroups,id)
	b.mu.Unlock()
}

type ResolverGroup struct{
	mu        sync.RWMutex
	id        string
	endpoints []string
	resolvers []*Resolver
}

// NewResolverGroup creates a new ResolverGroup with the given id.
func NewResolverGroup(id string) (*ResolverGroup, error) {
	return bldr.newResolverGroup(id)
}

func (e *ResolverGroup) Target(endpoint string) string {
	return Target(e.id,endpoint)
}

//endpoint://id/host endpoint
func Target(id string,endpoint string)string{
	return fmt.Sprintf("%s://%s/%s", scheme, id, endpoint)
}

func (e *ResolverGroup) SetEndpoints(endpoints []string) {
	addrs := epsToAddrs(endpoints...)
	e.mu.Lock()
	e.endpoints = endpoints
	for _,r := range e.resolvers{
		//grpc ClientConn newAddresss设置address
		r.cc.NewAddress(addrs)
	}
	e.mu.Unlock()
}

func (e *ResolverGroup)addResolver(r *Resolver){
	e.mu.Lock()
	addrs :=epsToAddrs(e.endpoints...)
	e.resolvers = append(e.resolvers,r)
	e.mu.Unlock()
	r.cc.NewAddress(addrs)
}

func (e *ResolverGroup) removeResolver(r *Resolver) {
	e.mu.Lock()
	for i, er := range e.resolvers {
		if er == r {
			e.resolvers = append(e.resolvers[:i], e.resolvers[i+1:]...)
			break
		}
	}
	e.mu.Unlock()
}

func (e *ResolverGroup)Close(){
	bldr.close(e.id)
}

// Resolver provides a resolver for a single etcd cluster, identified by name.
//包含单个EndpointID和grpc conn
type Resolver struct {
	endpointID string
	cc         resolver.ClientConn //grpc resolver ClientConn
	sync.RWMutex
}

func (*Resolver) ResolveNow(o resolver.ResolveNowOption) {}

func (r *Resolver) Close() {
	es, err := bldr.getResolverGroup(r.endpointID)
	if err != nil {
		return
	}
	es.removeResolver(r)
}

// TODO: use balancer.epsToAddrs
func epsToAddrs(eps ...string)(addrs []resolver.Address){
	addrs = make([]resolver.Address,0,len(eps))
	for _,ep := range eps{
		addrs = append(addrs,resolver.Address{Addr:ep})
	}
	return addrs
}

// ParseEndpoint endpoint parses an endpoint of the form
// (http|https)://<host>*|(unix|unixs)://<path>)
// and returns a protocol ('tcp' or 'unix'),
// host (or filepath if a unix socket),
// scheme (http, https, unix, unixs).
//Endpoint解析
func ParseEndpoint(endpoint string)(proto string,host string,scheme string){
	proto = "tpc"
	host = endpoint
	url,uerr := url.Parse(endpoint)
	if uerr != nil || !strings.Contains(endpoint,"://"){
		return proto,host,scheme
	}
	scheme = url.Scheme

	//strip scheme:// prefix since grpc dials by host
	host = url.Host
	switch url.Scheme{
	case "http", "https":
	case "unix", "unixs":
		proto = "unix"
		host = url.Host + url.Path
	default:
		proto,host = "",""
	}
	return proto,host,scheme
}
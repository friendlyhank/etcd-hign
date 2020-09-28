package embed

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/friendlyhank/etcd-hign/net/etcdserver"
	"github.com/friendlyhank/etcd-hign/net/etcdserver/api/etcdhttp"
	"github.com/friendlyhank/etcd-hign/net/etcdserver/api/rafthttp"
	"github.com/friendlyhank/etcd-hign/net/pkg/transport"
	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
)

type Etcd struct {
	Peers []*peerListener

	Server *etcdserver.EtcdServer

	cfg   Config
	stopc chan struct{}
	errc  chan error
}

type peerListener struct {
	net.Listener
	serve func() error
	close func(ctx context.Context) error
}

func StartEtcd(inCfg *Config) (e *Etcd, err error) {
	e = &Etcd{cfg: *inCfg}
	cfg := &e.cfg
	defer func() {
	}()

	//集群Listeners
	if e.Peers, err = configurePeerListeners(cfg); err != nil {
		return e, err
	}

	var (
		urlsmap types.URLsMap
		token   string
	)

	urlsmap, token, err = cfg.PeerURLsMapAndToken("etcd")
	if err != nil {
		return e, fmt.Errorf("error setting up initial cluster: %v", err)
	}

	srvcfg := etcdserver.ServerConfig{
		InitialPeerURLsMap:  urlsmap,
		InitialClusterToken: token,
	}

	//这里做的事情特别多
	//node start
	//transport start
	if e.Server, err = etcdserver.NewServer(srvcfg); err != nil {
		return e, err
	}

	//etcdserver start
	//raftNode start
	e.Server.Start()

	if err = e.servePeers(); err != nil {
		return e, err
	}

	return e, nil
}

// configurePeerListeners - 设置集群的监听
func configurePeerListeners(cfg *Config) (peers []*peerListener, err error) {
	peers = make([]*peerListener, len(cfg.LPUrls))
	defer func() {
	}()

	for i, u := range cfg.LPUrls {
		peers[i] = &peerListener{close: func(ctx context.Context) error { return nil }}
		peers[i].Listener, err = rafthttp.NewListener(u, &cfg.PeerTLSInfo)
		if err != nil {
			return nil, err
		}
		peers[i].close = func(ctx context.Context) error {
			return peers[i].Listener.Close()
		}
	}
	return peers, nil
}

// configureClientListeners -设置客户端的监听
func configureClientListeners(cfg *Config) (sctxs map[string]*serveCtx, err error) {
	sctxs = make(map[string]*serveCtx)
	for _, u := range cfg.LCUrls {
		sctx := newServeCtx()

		network := "tcp"
		addr := u.Host
		if u.Scheme == "unix" || u.Scheme == "unixs" {
			network = "unix"
			addr = u.Host + u.Path
		}
		sctx.network = network

		if sctx.l, err = net.Listen(network, addr); err != nil {
			return nil, err
		}
		// net.Listener will rewrite ipv4 0.0.0.0 to ipv6 [::], breaking
		// hosts that disable ipv6. So, use the address given by the user.
		sctx.addr = addr

		if network == "tcp" {
			if sctx.l, err = transport.NewKeepAliveListener(sctx.l, network, nil); err != nil {
				return nil, err
			}
		}
		sctxs[addr] = sctx
	}
	return sctxs, nil
}

//peers accept|read and write
func (e *Etcd) servePeers() (err error) {
	//http Handle
	ph := etcdhttp.NewPeerHandler(e.GetLogger(), e.Server)

	for _, p := range e.Peers {
		m := cmux.New(p.Listener)
		srv := &http.Server{
			Handler:     grpcHandlerFunc(nil, ph),
			ReadTimeout: 5 * time.Minute,
		}
		//http Handle业务逻辑
		go srv.Serve(m.Match(cmux.Any()))
		p.serve = func() error { return m.Serve() }
		p.close = func(ctx context.Context) error {
			return nil
		}
	}
	return nil
}

func (e *Etcd) serveClients() (err error) {

	return nil
}

func (e *Etcd) errHandler(err error) {
	select {
	case <-e.stopc:
		return
	default:
	}
	select {
	case <-e.stopc:
	case e.errc <- err:
	}
}

// GetLogger returns the logger.
func (e *Etcd) GetLogger() *zap.Logger {
	e.cfg.loggerMu.RLock()
	l := e.cfg.logger
	e.cfg.loggerMu.RUnlock()
	return l
}

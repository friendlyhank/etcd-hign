package embed

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/coreos/pkg/capnslog"
	runtimeutil "github.com/friendlyhank/etcd-hign/netmodule/pkg/runtime"
	"github.com/friendlyhank/etcd-hign/netmodule/pkg/transport"
	"github.com/friendlyhank/etcd-hign/netmodule/pkg/types"
	"github.com/friendlyhank/etcd-hign/netmodule/server/etcdserver"
	"github.com/friendlyhank/etcd-hign/netmodule/server/etcdserver/api/etcdhttp"
	"github.com/friendlyhank/etcd-hign/netmodule/server/etcdserver/api/rafthttp"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
)

var plog = capnslog.NewPackageLogger("github.com/friendlyhank/etcd-hign/netmodule", "embed")

const (
	// internal fd usage includes disk usage and transport usage.
	// To read/write snapshot, snap pkg needs 1. In normal case, wal pkg needs
	// at most 2 to read/lock/write WALs. One case that it needs to 2 is to
	// read all logs after some snapshot index, which locates at the end of
	// the second last and the head of the last. For purging, it needs to read
	// directory, so it needs 1. For fd monitor, it needs 1.
	// For transport, rafthttp builds two long-polling connections and at most
	// four temporary connections with each member. There are at most 9 members
	// in a cluster, so it should reserve 96.
	// For the safety, we set the total reserved number to 150.
	reservedInternalFDNum = 150
)

type Etcd struct {
	Peers   []*peerListener
	Clients []net.Listener
	// a map of contexts for the servers that serves client requests.
	sctxs map[string]*serveCtx

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
		if e == nil || err == nil {
			return
		}
	}()

	e.cfg.logger.Info(
		"configuring peer listeners",
		zap.Strings("listen-peer-urls", e.cfg.getLPURLs()),
	)
	//集群Listeners
	if e.Peers, err = configurePeerListeners(cfg); err != nil {
		return e, err
	}

	e.cfg.logger.Info(
		"configuring client listeners",
		zap.Strings("listen-clents-urls", e.cfg.getLCURLs()),
	)
	if e.sctxs, err = configureClientListeners(cfg); err != nil {
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
		Name:                cfg.Name,
		InitialPeerURLsMap:  urlsmap,
		InitialClusterToken: token,
	}

	//这里注意做的事情特别多
	//node start
	//transport start
	if e.Server, err = etcdserver.NewServer(srvcfg); err != nil {
		return e, err
	}

	//etcdserver start
	//raftNode start raftNode启动的时候会尝试去发送消息
	e.Server.Start()

	if err = e.servePeers(); err != nil {
		return e, err
	}
	if err = e.serveClients(); err != nil {
		return e, err
	}

	return e, nil
}

func (e *Etcd) Config() Config {
	return e.cfg
}

// configurePeerListeners - 设置集群的监听
func configurePeerListeners(cfg *Config) (peers []*peerListener, err error) {
	if err = updateCipherSuites(&cfg.PeerTLSInfo, cfg.CipherSuites); err != nil {
		return nil, err
	}
	if err = cfg.PeerSelfCert(); err != nil {
		if cfg.logger != nil {
			cfg.logger.Fatal("failed to get peer self-signed certs", zap.Error(err))
		} else {
			plog.Fatalf("could not get certs (%v)", err)
		}
	}

	if !cfg.PeerTLSInfo.Empty() {
		if cfg.logger != nil {
			cfg.logger.Info(
				"starting with peer TLS",
				zap.String("tls-info", fmt.Sprintf("%+v", cfg.PeerTLSInfo)),
				zap.Strings("cipher-suites", cfg.CipherSuites),
			)
		} else {
			plog.Infof("peerTLS: %s", cfg.PeerTLSInfo)
		}
	}

	peers = make([]*peerListener, len(cfg.LPUrls))
	defer func() {
		if err == nil {
			return
		}
		for i := range peers {
			//关闭连接
			for peers[i] != nil && peers[i].close != nil {
				cfg.logger.Warn(
					"closing peer listener",
					zap.String("address", cfg.LPUrls[i].String()),
					zap.Error(err),
				)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				peers[i].close(ctx)
				cancel()
			}
		}
	}()

	for i, u := range cfg.LPUrls {
		if u.Scheme == "http" {
			if !cfg.PeerTLSInfo.Empty() {
				if cfg.logger != nil {
					cfg.logger.Warn("scheme is HTTP while key and cert files are present; ignoring key and cert files", zap.String("peer-url", u.String()))
				} else {
					plog.Warningf("The scheme of peer url %s is HTTP while peer key/cert files are presented. Ignored peer key/cert files.", u.String())
				}
			}
			if cfg.PeerTLSInfo.ClientCertAuth {
				if cfg.logger != nil {
					cfg.logger.Warn("scheme is HTTP while --peer-client-cert-auth is enabled; ignoring client cert auth for this URL", zap.String("peer-url", u.String()))
				} else {
					plog.Warningf("The scheme of peer url %s is HTTP while client cert auth (--peer-client-cert-auth) is enabled. Ignored client cert auth for this url.", u.String())
				}
			}
		}
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

		//这个底层tcp是必须的，否则http无法连通
		for _, pl := range e.Peers {
			go func(l *peerListener) {
				u := l.Addr().String()
				if e.cfg.logger != nil {
					e.cfg.logger.Info(
						"serving peer traffic",
						zap.String("address", u),
					)
				} else {
					plog.Info("listening for peers on ", u)
				}
				e.errHandler(l.serve())
			}(pl)
		}
	}
	return nil
}

func configureClientListeners(cfg *Config) (sctxs map[string]*serveCtx, err error) {
	if err = updateCipherSuites(&cfg.ClientTLSInfo, cfg.CipherSuites); err != nil {
		return nil, err
	}
	if err = cfg.ClientSelfCert(); err != nil {
		if cfg.logger != nil {
			cfg.logger.Fatal("failed to get client self-signed certs", zap.Error(err))
		} else {
			plog.Fatalf("could not get certs (%v)", err)
		}
	}

	sctxs = make(map[string]*serveCtx)
	for _, u := range cfg.LCUrls {
		sctx := newServeCtx(cfg.logger)
		if u.Scheme == "http" || u.Scheme == "unix" {
			if !cfg.ClientTLSInfo.Empty() {
				if cfg.logger != nil {
					cfg.logger.Warn("scheme is HTTP while key and cert files are present; ignoring key and cert files", zap.String("client-url", u.String()))
				} else {
					plog.Warningf("The scheme of client url %s is HTTP while peer key/cert files are presented. Ignored key/cert files.", u.String())
				}
			}
			if cfg.ClientTLSInfo.ClientCertAuth {
				if cfg.logger != nil {
					cfg.logger.Warn("scheme is HTTP while --client-cert-auth is enabled; ignoring client cert auth for this URL", zap.String("client-url", u.String()))
				} else {
					plog.Warningf("The scheme of client url %s is HTTP while client cert auth (--client-cert-auth) is enabled. Ignored client cert auth for this url.", u.String())
				}
			}
		}
		if (u.Scheme == "https" || u.Scheme == "unixs") && cfg.ClientTLSInfo.Empty() {
			return nil, fmt.Errorf("TLS key/cert (--cert-file, --key-file) must be provided for client url %s with HTTPS scheme", u.String())
		}

		network := "tcp"
		addr := u.Host
		if u.Scheme == "unix" || u.Scheme == "unixs" {
			network = "unix"
			addr = u.Host + u.Path
		}
		sctx.network = network

		sctx.secure = u.Scheme == "https" || u.Scheme == "unixs"
		sctx.insecure = !sctx.secure
		if oldctx := sctxs[addr]; oldctx != nil {
			oldctx.secure = oldctx.secure || sctx.secure
			oldctx.insecure = oldctx.insecure || sctx.insecure
			//进来这里则是已经建立过连接
			continue
		}
		if sctx.l, err = net.Listen(network, addr); err != nil {
			return nil, err
		}
		// netmodule.Listener will rewrite ipv4 0.0.0.0 to ipv6 [::], breaking
		// hosts that disable ipv6. So, use the address given by the user.
		sctx.addr = addr

		if fdLimit, fderr := runtimeutil.FDLimit(); fderr == nil {
			if fdLimit <= reservedInternalFDNum {
				cfg.logger.Fatal(
					"file descriptor limit of etcd process is too low; please set higher",
					zap.Uint64("limit", fdLimit),
					zap.Int("recommended-limit", reservedInternalFDNum),
				)
			}
			sctx.l = transport.LimitListener(sctx.l, int(fdLimit-reservedInternalFDNum))
		}

		if network == "tcp" {
			if sctx.l, err = transport.NewKeepAliveListener(sctx.l, network, nil); err != nil {
				return nil, err
			}
		}

		defer func() {
			if err == nil {
				return
			}
			sctx.l.Close()
			if cfg.logger != nil {
				cfg.logger.Warn(
					"closing peer listener",
					zap.String("address", u.Host),
					zap.Error(err),
				)
			} else {
				plog.Info("stopping listening for client requests on ", u.Host)
			}
		}()
		sctxs[addr] = sctx
	}
	return sctxs, nil
}

//serveClients -生成http.handle 用于处理Client请求
func (e *Etcd) serveClients() (err error) {
	if !e.cfg.ClientTLSInfo.Empty() {
		if e.cfg.logger != nil {
			e.cfg.logger.Info(
				"starting with client TLS",
				zap.String("tls-info", fmt.Sprintf("%+v", e.cfg.ClientTLSInfo)),
				zap.Strings("cipher-suites", e.cfg.CipherSuites),
			)
		} else {
			plog.Infof("ClientTLS: %s", e.cfg.ClientTLSInfo)
		}
	}
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

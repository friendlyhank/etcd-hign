package embed

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/friendlyhank/etcd-hign/netmodule/pkg/logutil"
	"github.com/friendlyhank/etcd-hign/netmodule/pkg/tlsutil"

	"github.com/friendlyhank/etcd-hign/netmodule/pkg/transport"
	"github.com/friendlyhank/etcd-hign/netmodule/pkg/types"
	"go.uber.org/zap"
)

const (
	ClusterStateFlagNew      = "new"
	ClusterStateFlagExisting = "existing"

	DefaultName = "default"

	DefaultListenPeerURLs   = "http://localhost:2380"
	DefaultListenClientURLs = "http://localhost:2379"
)

var (
	DefaultInitialAdvertisePeerURLs = "http://localhost:2380"
	DefaultAdvertiseClientURLs      = "http://localhost:2379"
)

type Config struct {
	Name   string `json:"name"`
	Dir    string `json:"data-dir"` //数据目录的路径
	WalDir string `json:"wal-dir"`  //WAL文件专用目录

	// TickMs is the number of milliseconds between heartbeat ticks.
	// TODO: decouple tickMs and heartbeat tick (current heartbeat tick = 1).
	// make ticks a cluster wide configuration.
	TickMs     uint `json:"heartbeat-interval"` //选举定时心跳
	ElectionMs uint `json:"election-timeout"`   //选举超时

	LPUrls, LCUrls []url.URL
	ClientTLSInfo  transport.TLSInfo
	ClientAutoTLS  bool //是否自动生成Client TLS
	PeerTLSInfo    transport.TLSInfo
	PeerAutoTLS    bool //是否自动生成Peer TLS

	// CipherSuites is a list of supported TLS cipher suites between
	// client/server and peers. If empty, Go auto-populates the list.
	// Note that cipher suites are prioritized in the given order.
	CipherSuites []string `json:"cipher-suites"`

	InitialCluster      string `json:"initial-cluster"`
	InitialClusterToken string `json:"initial-cluster-token"`

	// Logger is logger options: currently only supports "zap".
	// "capnslog" is removed in v3.5.
	Logger string `json:"logger"`
	// LogLevel configures log level. Only supports debug, info, warn, error, panic, or fatal. Default 'info'.
	LogLevel string `json:"log-level"`

	loggerMu *sync.RWMutex
	logger   *zap.Logger
}

func NewConfig() *Config {
	lpurl, _ := url.Parse(DefaultListenPeerURLs)
	lcurl, _ := url.Parse(DefaultListenClientURLs)
	cfg := &Config{
		Name: DefaultName,

		TickMs:     100,  //选举定时时间
		ElectionMs: 1000, //选举超时时间

		LPUrls: []url.URL{*lpurl},
		LCUrls: []url.URL{*lcurl},

		InitialClusterToken: "etcd-cluster",

		loggerMu: new(sync.RWMutex),
		logger:   nil,
		Logger:   "zap",
		LogLevel: logutil.DefaultLogLevel,
	}
	cfg.InitialCluster = "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384"
	//Hank diff
	cfg.logger, _ = zap.NewProduction()
	return cfg
}

func (cfg *Config) PeerURLsMapAndToken(which string) (urlsmap types.URLsMap, token string, err error) {
	token = cfg.InitialClusterToken
	urlsmap, err = types.NewURLsMap(cfg.InitialCluster)
	return urlsmap, token, err
}

func updateCipherSuites(tls *transport.TLSInfo, ss []string) error {
	if len(tls.CipherSuites) > 0 && len(ss) > 0 {
		return fmt.Errorf("TLSInfo.CipherSuites is already specified (given %v)", ss)
	}
	if len(ss) > 0 {
		cs := make([]uint16, len(ss))
		for i, s := range ss {
			var ok bool
			cs[i], ok = tlsutil.GetCipherSuite(s)
			if !ok {
				return fmt.Errorf("unexpected TLS cipher suite %q", s)
			}
		}
		tls.CipherSuites = cs
	}
	return nil
}

func (cfg *Config) Validate() error {
	return nil
}

func (cfg *Config) ClientSelfCert() (err error) {
	if !cfg.ClientAutoTLS {
		return nil
	}
	if !cfg.ClientTLSInfo.Empty() {
		if cfg.logger != nil {
			cfg.logger.Warn("ignoring client auto TLS since certs given")
		} else {
			plog.Warningf("ignoring client auto TLS since certs given")
		}
		return nil
	}
	chosts := make([]string, len(cfg.LCUrls))
	for i, u := range cfg.LCUrls {
		chosts[i] = u.Host
	}
	cfg.ClientTLSInfo, err = transport.SelfCert(cfg.logger, filepath.Join(cfg.Dir, "fixtures", "client"), chosts)
	if err != nil {
		return err
	}
	return updateCipherSuites(&cfg.ClientTLSInfo, cfg.CipherSuites)
}

func (cfg *Config) PeerSelfCert() (err error) {
	if !cfg.PeerAutoTLS {
		return nil
	}
	if !cfg.PeerTLSInfo.Empty() {
		if cfg.logger != nil {
			cfg.logger.Warn("ignoring peer auto TLS since certs given")
		} else {
			plog.Warningf("ignoring peer auto TLS since certs given")
		}
		return nil
	}
	phosts := make([]string, len(cfg.LPUrls))
	for i, u := range cfg.LPUrls {
		phosts[i] = u.Host
	}
	cfg.PeerTLSInfo, err = transport.SelfCert(cfg.logger, filepath.Join(cfg.Dir, "fixtures", "peer"), phosts)
	if err != nil {
		return err
	}
	return updateCipherSuites(&cfg.PeerTLSInfo, cfg.CipherSuites)
}

func (cfg *Config) getLPURLs() (ss []string) {
	ss = make([]string, len(cfg.LPUrls))
	for i := range cfg.LPUrls {
		ss[i] = cfg.LPUrls[i].String()
	}
	return ss
}

func (cfg *Config) getLCURLs() (ss []string) {
	ss = make([]string, len(cfg.LCUrls))
	for i := range cfg.LCUrls {
		ss[i] = cfg.LCUrls[i].String()
	}
	return ss
}

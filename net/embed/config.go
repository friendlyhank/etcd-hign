package embed

import (
	"net/url"
	"sync"

	"go.etcd.io/etcd/pkg/logutil"

	"github.com/friendlyhank/etcd-hign/net/pkg/transport"
	"github.com/friendlyhank/etcd-hign/net/pkg/types"
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

	LPUrls, LCUrls []url.URL
	ClientTLSInfo  transport.TLSInfo
	ClientAutoTLS  bool //是否自动生成Client TLS
	PeerTLSInfo    transport.TLSInfo
	PeerAutoTLS    bool //是否自动生成Peer TLS

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
		Name:                DefaultName,
		LPUrls:              []url.URL{*lpurl},
		LCUrls:              []url.URL{*lcurl},
		InitialCluster:      "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
		InitialClusterToken: "etcd-cluster-1",

		loggerMu: new(sync.RWMutex),
		logger:   nil,
		Logger:   "zap",
		LogLevel: logutil.DefaultLogLevel,
	}
	//Hank diff
	cfg.logger, _ = zap.NewProduction()
	return cfg
}

func (cfg *Config) PeerURLsMapAndToken(which string) (urlsmap types.URLsMap, token string, err error) {
	token = cfg.InitialClusterToken
	urlsmap, err = types.NewURLsMap(cfg.InitialCluster)
	return urlsmap, token, err
}

func (cfg *Config) Validate() error {
	return nil
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

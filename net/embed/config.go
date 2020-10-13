package embed

import (
	"net/url"
	"sync"

	"github.com/friendlyhank/etcd-hign/net/pkg/transport"
	"github.com/friendlyhank/etcd-hign/net/pkg/types"
	"go.uber.org/zap"
)

const (
	DefaultName = "default"

	DefaultListenPeerURLs   = "http://localhost:2380"
	DefaultListenClientURLs = "http://localhost:2379"
)

type Config struct {
	Name           string `json:"name"`
	PeerTLSInfo    transport.TLSInfo
	LPUrls, LCUrls []url.URL

	InitialCluster      string `json:"initial-cluster"`
	InitialClusterToken string `json:"initial-cluster-token"`

	loggerMu *sync.RWMutex
	logger   *zap.Logger
}

func NewConfig() *Config {
	lpurl, _ := url.Parse(DefaultListenPeerURLs)
	lcurl, _ := url.Parse(DefaultListenClientURLs)
	cfg := &Config{
		Name:   DefaultName,
		LPUrls: []url.URL{*lpurl},
		LCUrls: []url.URL{*lcurl},
		//TODO Hank这个参数为啥解析不了
		InitialCluster:      "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
		InitialClusterToken: "etcd-cluster-1",
		loggerMu:            new(sync.RWMutex),
		logger:              nil,
	}
	return cfg
}

func (cfg *Config) PeerURLsMapAndToken(which string) (urlsmap types.URLsMap, token string, err error) {
	token = cfg.InitialClusterToken
	urlsmap, err = types.NewURLsMap(cfg.InitialCluster)
	return urlsmap, token, err
}

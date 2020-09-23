package embed

import (
	"net/url"
	"sync"

	"go.etcd.io/etcd/pkg/types"

	"github.com/friendlyhank/etcd-hign/net/pkg/transport"
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

		loggerMu: new(sync.RWMutex),
		logger:   nil,
	}
	return cfg
}

func (cfg *Config) PeerURLsMapAndToken() (urlsmap types.URLsMap, token string, err error) {

}

package embed

import (
	"github.com/friendlyhank/etcd-hign/net/pkg/transport"
	"net/url"
	"go.uber.org/zap"
	"sync"
)

const(

	DefaultName                  = "default"

	DefaultListenPeerURLs = "http://localhost:2380"
	DefaultListenClientURLs = "http://localhost:2379"
)

type Config struct {
	Name string `json:"name"`
	PeerTLSInfo    transport.TLSInfo
	LPUrls,LCUrls []url.URL

	loggerMu *sync.RWMutex
	logger   *zap.Logger
}

func NewConfig() *Config {
	lpurl,_ :=url.Parse(DefaultListenPeerURLs)
	lcurl,_ := url.Parse(DefaultListenClientURLs)
	cfg := &Config{
		Name: DefaultName,
		LPUrls:[]url.URL{*lpurl},
		LCUrls: []url.URL{*lcurl},

		loggerMu:   new(sync.RWMutex),
		logger:nil,
	}
	return cfg
}

package transport

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

type keepAliveConn interface {
	SetKeepAlive(bool) error
	SetKeepAlivePeriod(d time.Duration) error
}

func NewKeepAliveListener(l net.Listener, scheme string, tlscfg *tls.Config) (net.Listener, error) {
	if scheme == "https" {
		if tlscfg == nil {
			return nil, fmt.Errorf("cannot listen on TLS for given listener: KeyFile and CertFile are not presented")
		}
		return newTLSKeepaliveListener(l, tlscfg), nil
	}
	return &keepaliveListener{
		Listener: l,
	}, nil
}

type keepaliveListener struct{ net.Listener }

func (kln *keepaliveListener) Accept() (net.Conn, error) {
	return c, nil
}

type tlsKeepaliveListener struct {
	net.Listener
	config *tls.Config
}

func newTLSKeepaliveListener(inner net.Listener, config *tls.Config) net.Listener {
	l := &tlsKeepaliveListener{}
	l.Listener = inner
	l.config = config
	return l
}

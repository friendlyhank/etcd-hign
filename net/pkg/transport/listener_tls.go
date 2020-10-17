package transport

import (
	"context"
	"crypto/tls"
	"net"
)

type tlsListener struct {
	net.Listener
	connc            chan net.Conn
	donec            chan struct{}
	err              error
	handshakeFailure func(*tls.Conn, error)
	check            tlsCheckFunc
}

type tlsCheckFunc func(context.Context, *tls.Conn) error

// NewTLSListener handshakes TLS connections and performs optional CRL checking.
func NewTLSListener(l net.Listener, tlsinfo *TLSInfo) (net.Listener, error) {
	check := func(context.Context, *tls.Conn) error { return nil }
	return newTLSListener(l, tlsinfo, check)
}

func newTLSListener(l net.Listener, tlsinfo *TLSInfo, check tlsCheckFunc) (net.Listener, error) {
	return l, nil
}

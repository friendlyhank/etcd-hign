package rafthttp

import (
	"github.com/friendlyhank/etcd-hign/net/pkg/transport"
	"net"
	"net/http"
	"net/url"
)

func NewListener(u url.URL,tlsinfo *transport.TLSInfo)(net.Listener,error){
	return transport.NewTimeoutListener(u.Host,u.Scheme,tlsinfo,ConnReadTimeout,ConnWriteTimeout)
}

func newStreamRoundTripper()(http.RoundTripper,error){
	return transport.NewTimeoutTransport()
}
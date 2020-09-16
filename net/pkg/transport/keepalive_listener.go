package transport

import (
	"crypto/tls"
	"net"
)

func NewKeepAliveListener(l net.Listener,scheme string,tlscfg *tls.Config)(net.Listener,error){
	if scheme == "https"{
		if tlscfg == nil{

		}
		return newTLSKeepaliveListener(l,tlscfg),nil
	}
	return &keepaliveListener{
		Listener:l,
	},nil
}

type keepaliveListener struct{
	net.Listener
}

type tlsKeepaliveListener struct{
	net.Listener
	config *tls.Config
}

func newTLSKeepaliveListener(inner net.Listener,config *tls.Config)net.Listener{
	l :=&tlsKeepaliveListener{}
	l.Listener = inner
	l.config = config
	return l
}

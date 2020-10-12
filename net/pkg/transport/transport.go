package transport

import (
	"net"
	"net/http"
	"time"
)

type unixTransport struct{ *http.Transport }

//TODO HANK 搞懂这个是啥
func NewTransport()(*http.Transport,error){
	t := &http.Transport{
		Proxy:http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			// value taken from http.DefaultTransport
			KeepAlive: 30 * time.Second,
		}).Dial,
	}

	dialer := (&net.Dialer{
		KeepAlive: 30 * time.Second,
	})
	dial := func(net, addr string) (net.Conn, error) {
		return dialer.Dial("unix", addr)
	}

	tu := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                dial,
	}

	ut :=&unixTransport{tu}

	t.RegisterProtocol("unix", ut)
	t.RegisterProtocol("unixs",ut)

	return t,nil
}

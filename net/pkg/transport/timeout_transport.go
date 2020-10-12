package transport

import (
	"net"
	"net/http"
	"time"
)

// NewTimeoutTransport returns a transport created using the given TLS info.
// If read/write on the created connection blocks longer than its time limit,
// it will return timeout error.
// If read/write timeout is set, transport will not be able to reuse connection.
func NewTimeoutTransport() (*http.Transport, error) {
	tr, err := NewTransport()
	if err != nil{
		return nil,err
	}

	tr.Dial = (&rwTimeoutDialer{
		Dialer: net.Dialer{
			KeepAlive: 30 * time.Second,
		},
	}).Dial
	return tr,nil
}

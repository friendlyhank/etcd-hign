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
func NewTimeoutTransport(info TLSInfo, dialtimeoutd, rdtimeoutd, wtimeoutd time.Duration) (*http.Transport, error) {
	tr, err := NewTransport(info, dialtimeoutd)
	if err != nil {
		return nil, err
	}

	if rdtimeoutd != 0 || wtimeoutd != 0 {
		tr.MaxIdleConnsPerHost = 1
	} else {
		tr.MaxIdleConnsPerHost = 1024
	}

	tr.Dial = (&rwTimeoutDialer{
		Dialer: net.Dialer{
			Timeout:   dialtimeoutd,
			KeepAlive: 30 * time.Second,
		},
		rdtimeoutd: rdtimeoutd,
		wtimeoutd:  wtimeoutd,
	}).Dial
	return tr, nil
}

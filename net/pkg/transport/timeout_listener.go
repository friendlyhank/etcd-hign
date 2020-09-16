package transport

import (
	"net"
	"time"
)

// NewTimeoutListener returns a listener that listens on the given address
// If read/wtite on the accepted connection blocks longer than its time limit
// it will return timeout error.
func NewTimeoutListener(addr string,scheme string,tlsinfo *TLSInfo,rdtimeoutd,wtimeoutd time.Duration)(net.Listener,error){
	ln,err := newListener(addr,scheme)
	if err != nil{
		return nil,err
	}
	return ln,nil
}

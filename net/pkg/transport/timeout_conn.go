package transport

import (
	"net"
	"time"
)

type timeoutConn struct {
	net.Conn
	wtimeoutd  time.Duration
	rdtimeoutd time.Duration
}

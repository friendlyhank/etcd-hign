package transport

import (
	"net"
	"net/http"
	"testing"
)

func TestNewKeepAliveListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unexpected listen error: %v", err)
	}
	ln, err = NewKeepAliveListener(ln, "http", nil)
	if err != nil {
		t.Fatalf("unexpected NewKeepAliveListener error: %v", err)
	}

	go http.Get("http://" + ln.Addr().String())
	conn, err := ln.Accept()
	if err != nil {
		t.Fatalf("unexpected Accept error: %v", err)
	}
	conn.Close()
	ln.Close()
}

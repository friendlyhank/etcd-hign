package transport

import (
	"net"
	"testing"
	"time"
)

func TestReadWriteTimeoutDialer(t *testing.T) {
	stop := make(chan struct{})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unexpected listen error:%v", err)
	}
	ts := testBlockingServer{ln, 2, stop}
	go ts.Start(t)

	d := rwTimeoutDialer{
		wtimeoutd:  10 * time.Microsecond,
		rdtimeoutd: 10 * time.Microsecond,
	}
	conn, err := d.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("unexpected dial error:%v", err)
	}
	defer conn.Close()

	//fill the socket buffer
	data := make([]byte, 5*1024*1024)
	done := make(chan struct{})
	go func() {
		_, err = conn.Write(data)
		done <- struct{}{}
	}()

	select {
	case <-done:
	}
}

type testBlockingServer struct {
	ln   net.Listener
	n    int
	stop chan struct{}
}

func (ts *testBlockingServer) Start(t *testing.T) {
	for i := 0; i < ts.n; i++ {
		conn, err := ts.ln.Accept()
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
	}
	<-ts.stop
}

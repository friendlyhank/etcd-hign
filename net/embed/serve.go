package embed

import (
	"context"
	"net"
)

type serveCtx struct{
	l net.Listener
	addr string
	network string

	ctx context.Context
	cancel context.CancelFunc
}

func newServeCtx() *serveCtx {
	ctx,cancel := context.WithCancel(context.Background())
	return &serveCtx{
		ctx: ctx,
		cancel: cancel,
	}
}

package embed

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

type serveCtx struct {
	l       net.Listener
	addr    string
	network string

	ctx    context.Context
	cancel context.CancelFunc
}

func newServeCtx() *serveCtx {
	ctx, cancel := context.WithCancel(context.Background())
	return &serveCtx{
		ctx:    ctx,
		cancel: cancel,
	}
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	if otherHandler == nil {

	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otherHandler.ServeHTTP(w, r)
	})
}

package embed

import (
	"context"
	"net"
	"net/http"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

type serveCtx struct {
	lg      *zap.Logger
	l       net.Listener
	addr    string
	network string

	ctx    context.Context
	cancel context.CancelFunc
}

func newServeCtx(lg *zap.Logger) *serveCtx {
	ctx, cancel := context.WithCancel(context.Background())
	if lg == nil {
		lg = zap.NewNop()
	}
	return &serveCtx{
		lg:     lg,
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

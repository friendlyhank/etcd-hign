package etcdmain

import (
	"fmt"
	"os"

	"github.com/friendlyhank/etcd-hign/net/embed"
	"github.com/friendlyhank/etcd-study/pkg/osutil"
	"go.uber.org/zap"
)

func startEtcdOrProxyV2() {
	cfg := newConfig()

	var err error

	//启动日志组件
	lg := cfg.ec.GetLogger()
	if lg == nil {
		var zapError error
		lg, zapError = zap.NewProduction()
		if zapError != nil {
			fmt.Printf("error creating zap logger %v", zapError)
			os.Exit(1)
		}
	}
	if err != nil {

	}

	var stopped <-chan struct{}
	var errc <-chan error
	stopped, errc, err = startEtcd(&cfg.ec)

	select {
	case lerr := <-errc:
		lg.Fatal("listener failed", zap.Error(lerr))
	case <-stopped:
	}

	osutil.Exit(0)
}

func startEtcd(cfg *embed.Config) (<-chan struct{}, <-chan error, error) {
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, nil, err
	}
	select {
	case <-e.Server.ReadyNotify():
	}
	return nil, nil, nil
}

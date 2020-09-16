package etcdmain

import (
	"fmt"
	"github.com/friendlyhank/etcd-hign/net/embed"
	"go.uber.org/zap"
	"os"
)

func startEtcdOrProxyV2(){
	cfg := newConfig()

	//启动日志组件
	lg := cfg.ec.GetLogger()
	if lg == nil{
		var zapError error
		lg, zapError = zap.NewProduction()
		if zapError != nil {
			fmt.Printf("error creating zap logger %v", zapError)
			os.Exit(1)
		}
	}
	_,_,_ =startEtcd(&cfg.ec)
}

func startEtcd(cfg *embed.Config)(<-chan struct{},<-chan error,error){
	embed.StartEtcd(cfg)
	return nil,nil,nil
}

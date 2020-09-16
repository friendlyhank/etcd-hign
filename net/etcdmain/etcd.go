package etcdmain

import "github.com/friendlyhank/etcd-hign/net/embed"

func startEtcdOrProxyV2(){
	cfg := newConfig()
	_,_,_ =startEtcd(&cfg.ec)
}

func startEtcd(cfg *embed.Config)(<-chan struct{},<-chan error,error){
	embed.StartEtcd(cfg)
	return nil,nil,nil
}

package etcdmain

import "github.com/friendlyhank/etcd-hign/net/embed"

func startEtcdOrProxyV2(){
	_,_,_ =startEtcd()
}

func startEtcd()(<-chan struct{},<-chan error,error){
	embed.StartEtcd()
	return nil,nil,nil
}

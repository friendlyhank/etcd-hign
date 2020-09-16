package embed

import (
)

type Etcd struct{
	cfg Config
}

type peerListener struct{
}

func StartEtcd(inCfg *Config)(e *Etcd,err error){
	e = &Etcd{cfg:*inCfg}
	cfg := &e.cfg
	defer func() {
	}()

	//集群Listeners
	configurePeerListeners(cfg)

	//客户端Listeners
	configureClientListeners(cfg)
	return e,nil
}

// configurePeerListeners - 设置集群的监听
func configurePeerListeners(cfg *Config)(peers []*peerListener,err error){
	peers =make([]*peerListener,len(cfg.LPUrls))
	defer func() {
	}()

	for i,u := range cfg.LPUrls{

	}
	return peers,nil
}

// configureClientListeners -设置客户端的监听
func configureClientListeners(cfg *Config){

}

func (e *Etcd)servePeers()(err error){

}

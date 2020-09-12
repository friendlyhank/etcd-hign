package embed

import(
	"net"
)

type Etcd struct {
	Peers   []*peerListener //集群Listener
	Clients []net.Listener  //客户端Listener
	// a map of contexts for the servers that serves client requests.
	sctxs map[string]*serveCtx //用map去装Peers或Clients listener


}

type peerListener struct {
	net.Listener
	serve func() error
}

func StartEtcd(inCfg *Config) (e *Etcd, err error) {

}

func configureClientListeners(cfg *Config) (sctxs map[string]*serveCtx, err error) {
	sctxs = make(map[string]*serveCtx)
	for _, u := range cfg.LCUrls {

	}
}

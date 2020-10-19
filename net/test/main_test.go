package test

import (
	"os"
	"testing"

	"github.com/friendlyhank/etcd-hign/net/etcdmain"
)

//生产环境中端口应该是一样的，IP不同
//启动测试服务
func StartInfraServer(args []string) {
	etcdmain.Main() //服务端主入口
}

/*====================================单节点配置===============================================*/
func TestSingleEtcdMain(t *testing.T) {
	os.Args = []string{"etcd-test"}
	StartInfraServer(os.Args)
}

/*====================================集群配置===============================================*/
func TestInfraOEtcdMain(t *testing.T) {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra0",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2380",
		"--listen-peer-urls", "http://127.0.0.1:2380",
		"--listen-client-urls", "http://127.0.0.1:2379",
		"--advertise-client-urls", "http://127.0.0.1:2379",
		"--initial-cluster-token", "etcd-cluster-1",
		"--initial-cluster", "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
		"--initial-cluster-state", "new"}
	StartInfraServer(os.Args)
}

func TestInfra1EtcdMain(t *testing.T) {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra1",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2382",
		"--listen-peer-urls", "http://127.0.0.1:2382",
		"--listen-client-urls", "http://127.0.0.1:2381",
		"--advertise-client-urls", "http://127.0.0.1:2381",
		"--initial-cluster-token", "etcd-cluster-1",
		"--initial-cluster", "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
		"--initial-cluster-state", "new"}
	StartInfraServer(os.Args)
}

func TestInfra2EtcdMain(t *testing.T) {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra2",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2384",
		"--listen-peer-urls", "http://127.0.0.1:2384",
		"--listen-client-urls", "http://127.0.0.1:2383",
		"--advertise-client-urls", "http://127.0.0.1:2383",
		"--initial-cluster-token", "etcd-cluster-1",
		"--initial-cluster", "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
		"--initial-cluster-state", "new"}
	StartInfraServer(os.Args)
}

/*====================================集群服务发现===============================================*/
func TestInfraODiscoverEtcdMain(t *testing.T) {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infraO",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2382",
		"--listen-peer-urls", "http://127.0.0.1:2382",
		"--listen-client-urls", "http://127.0.0.1:2381",
		"--advertise-client-urls", "http://127.0.0.1:2381",
		"--discovery", "https://discovery.etcd.io/f0be1fdc930b1d1a495bb99544a4d2b7",
	}
	StartInfraServer(os.Args)
}

func TestInfra1DiscoverEtcdMain(t *testing.T) {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra1",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2384",
		"--listen-peer-urls", "http://127.0.0.1:2384",
		"--listen-client-urls", "http://127.0.0.1:2383",
		"--advertise-client-urls", "http://127.0.0.1:2383",
		"--discovery", "https://discovery.etcd.io/f0be1fdc930b1d1a495bb99544a4d2b7",
	}
	StartInfraServer(os.Args)
}

func TestInfra2DiscoverEtcdMain(t *testing.T) {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra2",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2386",
		"--listen-peer-urls", "http://127.0.0.1:2386",
		"--listen-client-urls", "http://127.0.0.1:2385",
		"--advertise-client-urls", "http://127.0.0.1:2385",
		"--discovery", "https://discovery.etcd.io/f0be1fdc930b1d1a495bb99544a4d2b7",
	}
	StartInfraServer(os.Args)
}

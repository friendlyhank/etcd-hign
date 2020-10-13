package test

import (
	"net/http"
	"os"
	"testing"

	"github.com/friendlyhank/etcd-hign/net/etcdmain"
)

/*====================================单节点配置===============================================*/
func TestSingleEtcdMain(t *testing.T) {
	os.Args = []string{"etcd-test"}
	etcdmain.Main()
}

/*====================================集群配置===============================================*/
//生产环境中端口应该是一样的，IP不同
//启动测试服务1
func StartInfraOServer() {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra0",
		"--listen-peer-urls", "http://127.0.0.1:2380",
		"--listen-client-urls", "http://127.0.0.1:2379",
		"--initial-cluster", "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
	}
	etcdmain.Main() //服务端主入口
}

func StartInfra1Server() {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra1",
		"--listen-peer-urls", "http://127.0.0.1:2382",
		"--listen-client-urls", "http://127.0.0.1:2381",
		"--initial-cluster", "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
	}
	etcdmain.Main() //服务端主入口
}

func StartInfra2Server() {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra2",
		"--listen-peer-urls", "http://127.0.0.1:2384",
		"--listen-client-urls", "http://127.0.0.1:2383",
		"--initial-cluster", "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
	}
	etcdmain.Main() //服务端主入口
}

func TestInfraOEtcdMain(t *testing.T) {
	StartInfraOServer()
}

func TestInfra1EtcdMain(t *testing.T) {
	StartInfra1Server()
}

func TestInfra2EtcdMain(t *testing.T) {
	StartInfra2Server()
}

/*====================================Http请求===============================================*/
func TestSendHttpGet(t *testing.T) {
	//不知道从哪里的数据源会发送一个http请求
	http.Get("http://127.0.0.1:2380/raft/stream/message/19ac17627e3e396f")
}

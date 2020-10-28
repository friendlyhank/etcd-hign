package test

import (
	"crypto/x509"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"go.etcd.io/etcd/pkg/transport"

	"github.com/friendlyhank/etcd-hign/net/etcdmain"
)

//生成https签名校验
func createSelfCert(host string) (*transport.TLSInfo, func(), error) {
	return createSelfCertEx(host)
}

func createSelfCertEx(host string, additionalUsages ...x509.ExtKeyUsage) (*transport.TLSInfo, func(), error) {
	d, terr := ioutil.TempDir("", "etcd-test-tls-")
	if terr != nil {
		return nil, nil, terr
	}
	info, err := transport.SelfCert(nil, d, []string{host}, additionalUsages...)
	if err != nil {
		return nil, nil, err
	}
	return &info, func() { os.RemoveAll(d) }, nil
}

//生产环境中端口应该是一样的，IP不同
//启动测试服务
func StartInfraServer() {
	etcdmain.Main() //服务端主入口
}

/*====================================单节点配置===============================================*/
func TestSingleEtcdMain(t *testing.T) {
	os.Args = []string{"etcd-test"}
	StartInfraServer()
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
		"--initial-cluster-state", "new",
		//方便调试
		"--logger=zap",
		"--log-level=error", //日志等级 debug, info, warn, error, panic, or fatal
	}
	StartInfraServer()
}

func TestInfra1EtcdMain(t *testing.T) {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra1",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2382",
		"--listen-peer-urls", "http://127.0.0.1:2382",
		"--listen-client-urls", "http://127.0.0.1:2381",
		"--advertise-client-urls", "http://127.0.0.1:2381",
		"--initial-cluster-token", "etcd-cluster-1",
		"--initial-cluster", "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
		"--initial-cluster-state", "new",
		//方便调试
		"--logger=zap",
		"--log-level=error", //日志等级 debug, info, warn, error, panic, or fatal
	}
	StartInfraServer()
}

func TestInfra2EtcdMain(t *testing.T) {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra2",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2384",
		"--listen-peer-urls", "http://127.0.0.1:2384",
		"--listen-client-urls", "http://127.0.0.1:2383",
		"--advertise-client-urls", "http://127.0.0.1:2383",
		"--initial-cluster-token", "etcd-cluster-1",
		"--initial-cluster", "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
		"--initial-cluster-state", "new",
		//方便调试
		"--logger=zap",
		"--log-level=error", //日志等级 debug, info, warn, error, panic, or fatal
	}
	StartInfraServer()
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
	StartInfraServer()
}

func TestInfra1DiscoverEtcdMain(t *testing.T) {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra1",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2384",
		"--listen-peer-urls", "http://127.0.0.1:2384",
		"--listen-client-urls", "http://127.0.0.1:2383",
		"--advertise-client-urls", "http://127.0.0.1:2383",
		"--discovery", "https://discovery.etcd.io/f0be1fdc930b1d1a495bb99544a4d2b7",
	}
	StartInfraServer()
}

func TestInfra2DiscoverEtcdMain(t *testing.T) {
	os.Args = []string{"etcd-3.3.12-test", "--name", "infra2",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2386",
		"--listen-peer-urls", "http://127.0.0.1:2386",
		"--listen-client-urls", "http://127.0.0.1:2385",
		"--advertise-client-urls", "http://127.0.0.1:2385",
		"--discovery", "https://discovery.etcd.io/f0be1fdc930b1d1a495bb99544a4d2b7",
	}
	StartInfraServer()
}

type securityConfig struct {
	CertFile      string `json:"cert-file"`
	KeyFile       string `json:"key-file"`
	CertAuth      bool   `json:"client-cert-auth"`
	TrustedCAFile string `json:"trusted-ca-file"`
	AutoTLS       bool   `json:"auto-tls"`
}

func TestInfraOEtcdMainWithTLS(t *testing.T) {
	peerTls, peerDel, peerErr := createSelfCert("127.0.0.1:2380")
	clientTls, clientDel, clientErr := createSelfCert("127.0.0.1:2379")
	if peerErr != nil {
		t.Fatalf("unable to create peercert: %v", peerErr)
	}
	if clientErr != nil {
		t.Fatalf("unable to create clientcert: %v", clientErr)
	}
	defer peerDel()
	defer clientDel()

	if runtime.GOOS == "windows" {
		clientTls.CertFile = strings.ReplaceAll(clientTls.CertFile, "\\", "/")
		clientTls.KeyFile = strings.ReplaceAll(clientTls.KeyFile, "\\", "/")
		peerTls.CertFile = strings.ReplaceAll(peerTls.CertFile, "\\", "/")
		peerTls.KeyFile = strings.ReplaceAll(peerTls.KeyFile, "\\", "/")
	}

	os.Args = []string{"etcd-3.3.12-test", "--name", "infra0",
		"--initial-advertise-peer-urls", "http://127.0.0.1:2380",
		"--listen-peer-urls", "http://127.0.0.1:2380",
		"--listen-client-urls", "http://127.0.0.1:2379",
		"--advertise-client-urls", "http://127.0.0.1:2379",
		"--initial-cluster-token", "etcd-cluster-1",
		"--initial-cluster", "infra0=http://127.0.0.1:2380,infra1=http://127.0.0.1:2382,infra2=http://127.0.0.1:2384",
		"--initial-cluster-state", "new",
		//"trusted-ca-file" //受信任的证书办法机构
		//"peer-trusted-ca-file" //perr受信任的证书办法机构
		"--cert-file", clientTls.CertFile,
		"--key-file", clientTls.KeyFile,
		"--peer-cert-file", peerTls.CertFile,
		"--peer-key-file", peerTls.KeyFile,
		"--peer-client-cert-auth=true", //peer cert校验
		"--client-cert-auth=true",      //客户端cert校验
	}
	StartInfraServer()
}

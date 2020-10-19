package transport

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"go.uber.org/zap"
)

// createSelfCert 生成TLSInfo
func createSelfCert(hosts ...string) (*TLSInfo, func(), error) {
	return createSelfCertEx("127.0.0.1")
}

func createSelfCertEx(host string, additionalUsages ...x509.ExtKeyUsage) (*TLSInfo, func(), error) {
	d, terr := ioutil.TempDir("", "etcd-test-tls-")
	if terr != nil {
		return nil, nil, terr
	}
	info, err := SelfCert(zap.NewExample(), d, []string{host + ":0"}, additionalUsages...)
	if err != nil {
		return nil, nil, err
	}
	return &info, func() { os.RemoveAll(d) }, nil
}

// TestNewListenerTLSInfo tests that NewListener with valid TLSInfo returns
// a TLS listener that accepts TLS connections.
func TestNewListenerTLSInfo(t *testing.T) {
	tlsInfo, del, err := createSelfCert()
	if err != nil {
		t.Fatalf("unable to create cert:%v", err)
	}
	defer del()
	testNewListenerTLSInfoAccept(t, *tlsInfo)
}

func testNewListenerTLSInfoAccept(t *testing.T, tlsInfo TLSInfo) {
	ln, err := NewListener("127.0.0.1:0", "https", &tlsInfo)

	if err != nil {
		t.Fatalf("unexpected NewListener error: %v", err)
	}
	defer ln.Close()

	//TLS不验证服务器身份即访问
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	cli := http.Client{Transport: tr}
	go cli.Get("https://" + ln.Addr().String())

	conn, err := ln.Accept()
	if err != nil {
		t.Fatalf("unexpected Accept error: %v", err)
	}
	defer conn.Close()
	if _, ok := conn.(*tls.Conn); !ok {
		t.Error("failed to accept *tls.Conn")
	}
}

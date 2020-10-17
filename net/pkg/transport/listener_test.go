package transport

import (
	"crypto/x509"
	"io/ioutil"
	"os"
	"testing"

	"go.uber.org/zap"
)

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
}

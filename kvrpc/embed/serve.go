package embed

import(
	"net"
)

//服务客户端的请求
type serveCtx struct {
	l       net.Listener
	addr    string
	network string

	secure   bool //是否 https证书校验
	insecure bool
}

package client3

import(
	"time"
	)

type Config struct{
	// Endpoints is a list of URLs.
	Endpoints []string `json:"endpoints"`

	// DialTimeout is the timeout for failing to establish a connection.
	DialTimeout time.Duration `json:"dial-timeout"`

	//grpc keepAlive
	DialKeepAliveTime time.Duration `json:"dial-keep-alive-time"`

	// DialKeepAliveTimeout is the time that the client waits for a response for the
	// keep-alive probe. If the response is not received in this time, the connection is closed.
	DialKeepAliveTimeout time.Duration `json:"dial-keep-alive-timeout"`
	
	// MaxCallSendMsgSize is the client-side request send limit in bytes.
	// If 0, it defaults to 2.0 MiB (2 * 1024 * 1024).
	// Make sure that "MaxCallSendMsgSize" < server-side default send/recv limit.
	// ("--max-request-bytes" flag to etcd or "embed.Config.MaxRequestBytes").
	MaxCallSendMsgSize int //grpc最大请求msg

	// MaxCallRecvMsgSize is the client-side response receive limit.
	// If 0, it defaults to "math.MaxInt32", because range response can
	// easily exceed request send limits.
	// Make sure that "MaxCallRecvMsgSize" >= server-side default send/recv limit.
	// ("--max-request-bytes" flag to etcd or "embed.Config.MaxRequestBytes").
	MaxCallRecvMsgSize int//grpc最大接收

	// PermitWithoutStream when set will allow client to send keepalive pings to server without any active streams(RPCs).
	PermitWithoutStream bool `json:"permit-without-stream"`
}

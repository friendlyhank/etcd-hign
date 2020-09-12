package etcdserver

type ServerConfig struct {
	Name           string

	// MaxRequestBytes is the maximum request size to send over raft.
	//grpc最大请求bytes
	MaxRequestBytes uint
}

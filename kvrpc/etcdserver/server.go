package etcdserver

import(
	"google.golang.org/grpc"
	pb "github.com/friendlyhank/etcd-hign/kvrpc/etcdserver/etcdserverpb"
	"context"
	"crypto/tls"
)

const (
	grpcOverheadBytes = 512 * 1024
)

type RaftKV interface {
	Put(ctx context.Context, r *pb.PutRequest) (*pb.PutResponse, error)
}

type EtcdServer struct {
	Cfg ServerConfig
}

func NewKVServer() pb.KVServer {
	return &EtcdServer{}
}

func (s *EtcdServer)Put(ctx context.Context, r *pb.PutRequest) (*pb.PutResponse, error){
	return &pb.PutResponse{Key:[]byte("foo")},nil
}

func (s *EtcdServer)Server(tls *tls.Config, gopts ...grpc.ServerOption) *grpc.Server {
	var opts []grpc.ServerOption
	opts = append(opts, grpc.MaxRecvMsgSize(int(s.Cfg.MaxRequestBytes+grpcOverheadBytes)))

	//new grpc Server实例
	grpcServer := grpc.NewServer(append(opts, gopts...)...)

	pb.RegisterKVServer(grpcServer, NewKVServer())

	return grpcServer
}
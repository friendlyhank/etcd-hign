package client3

import(
	"context"
	"google.golang.org/grpc"
	pb "hank.com/etcd-3.3.12-hign/kvrpc/etcdserver/etcdserverpb"
)

type (
	PutResponse pb.PutResponse
)

type OpResponse struct{
	put *PutResponse
}

func(op OpResponse)Put()*PutResponse{return op.put}

func (resp *PutResponse)OpResponse()OpResponse{
	return OpResponse{put:resp}
}

type KV interface {
	Put(ctx context.Context, key, val string, opts ...OpOption) (*PutResponse, error)
}

func NewKV(c *Client)KV{
	return &kv{remote:pb.NewKVClient(c.conn)}
}

type kv struct{
	remote pb.KVClient
	callOpts []grpc.CallOption
}

func (kv *kv)Put(ctx context.Context,key string,val string,opts ...OpOption)(*PutResponse,error){
	r,err := kv.Do(ctx,OpPut(key,val,opts...))
	return r.put,toErr(ctx,err)
}

func (kv *kv)Do(ctx context.Context,op Op)(OpResponse,error){
	var err error
	switch op.t {
	case tPut:
		var resp *pb.PutResponse
		r:=&pb.PutRequest{}
		resp,err = kv.remote.Put(ctx,r,kv.callOpts...)
		if err == nil{
			return OpResponse{put:(*PutResponse)(resp)},nil
		}
	default:
		panic("Unknown op")
	}
	return OpResponse{},toErr(ctx,err)
}

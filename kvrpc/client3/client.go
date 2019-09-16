package client3

import(
	"google.golang.org/grpc"
	"context"
	)

type Client struct{
	KV

	conn *grpc.ClientConn //grpc链接
}

func New(cfg Config)(*Client,error){
	return newClient(&cfg)
}

//根据Config去new Client
func newClient(cfg *Config)(*Client,error){
	if cfg == nil{
		cfg =&Config{}
	}

	client := &Client{

	}

	client.KV = NewKV(client)

	return client,nil
}

func toErr(ctx context.Context, err error) error {
	if err == nil{
		return nil
	}

	return err
}

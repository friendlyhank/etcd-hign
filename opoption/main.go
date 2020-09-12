package main

import (
	"context"
	"github.com/friendlyhank/etcd-hign/opoption/embed"
)

func main(){

	//(1)test Op 这样可以借助OpOption做很多事情
	k := &embed.Kv{}
	k.Put(context.Background(),"key","value",embed.WithLease(123456))


	//(2)提供入口,api接口

	//设置conn之后，拿到调取api的接口
	client := &embed.Client{Conn:nil}

	client.KV = embed.NewKV(client)
	k.Put(context.Background(),"key","value",embed.WithLease(123456))

	//(3)总分总形式，方法聚合有可以分离
	//具体从案例中读取

}
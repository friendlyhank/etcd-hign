package main

import (
	"context"
	"hank.com/etcd-3.3.12-hign/opoption/embed"
)

func main(){
	k := &embed.Kv{}

	k.Put(context.Background(),"key","value",embed.WithLease(123456))
}
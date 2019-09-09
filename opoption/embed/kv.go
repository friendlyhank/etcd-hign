package embed

import (
	"context"
	"net"
)

type KV interface{
	Put(ctx context.Context,key,val string,opts ...OpOption)(*PutResponse, error)
	Get(ctx context.Context, key string, opts ...OpOption) (*GetResponse, error)
	Delete(ctx context.Context, key string, opts ...OpOption) (*DeleteResponse, error)
	Do(ctx context.Context, op Op) (OpResponse, error)
}

type Lease interface {
}

type Watcher interface {
}

//set pb
type PutResponse struct{}
type GetResponse struct{}
type DeleteResponse struct{}

/*总分总形式，方法聚合有可以分离
Put、Get、Delete都是用Do,结构返回统一用OpResponse,又可以获取具体的，如Put put := op.Put()
*/
type OpResponse struct{
	put *PutResponse
	get *GetResponse
	del *DeleteResponse
}

func (op OpResponse)Put()*PutResponse{return op.put}
func (op OpResponse)Get()*GetResponse{return op.get}
func (op OpResponse)Del()*DeleteResponse{return op.del}

func (resp *PutResponse)OpResponse()OpResponse{
	return OpResponse{put:resp}
}

func (resp *GetResponse)OpResponse()OpResponse{
	return OpResponse{get:resp}
}

func (resp *DeleteResponse)OpResponse()OpResponse{
	return OpResponse{del:resp}
}

//Clinet 结构体
type Client struct{
	KV
	Lease
	Watcher
	Conn *net.Conn
}

/*提供入口,api接口
*/
func NewKV(c *Client) KV {
	api := &Kv{Conn:c.Conn}
	return api
}

type Kv struct{
	Conn *net.Conn
}

func (kv *Kv)Get(ctx context.Context, key string, opts ...OpOption) (*GetResponse, error){
	return nil,nil
}

func (kv *Kv) Put(ctx context.Context, key, val string, opts ...OpOption)(*PutResponse, error) {
	//注意不是引用类型传入了引用类型的函数
	op := Op{t:tPut}
	op.applyOpts(opts)

	r,err := kv.Do(ctx,op)
	return r.put,err
}

func (kv *Kv)Delete(ctx context.Context, key string, opts ...OpOption) (*DeleteResponse, error){
	return nil,nil
}

func (kv *Kv)Do(ctx context.Context, op Op) (OpResponse, error){
	//判断类型
	switch op.t {
	case tPut:
		return OpResponse{put:&PutResponse{}},nil
	}

	return OpResponse{},nil
}
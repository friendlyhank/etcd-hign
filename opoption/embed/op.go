package embed

import "fmt"

/**
 *op etcd比较常见的写法
 *适用于可选条件，且参数比较多的时候可以用
 *但不适用于同时用多个方法(可读性差)
 */

type opType int

const (
	// A default Op has opType 0, which is invalid.
	tRange opType = iota + 1
	tPut
	tDeleteRange
	tTxn
)

type Op struct{
	t   opType //操作类型
	LeaseID int64
}

type OpOption func(*Op)

func WithLease(leaseID int64) OpOption{
	return  func(o *Op){o.LeaseID = leaseID}
}

func (o *Op)applyOpts(ops []OpOption){
	fmt.Println(o)
	for _,opOption := range ops{
		opOption(o)
	}
}



package embed

/**
 *op etcd比较常见的写法
 *适用于可选条件，且参数比较多的时候可以用
 *但不适用于同时用多个方法(可读性差)
 */
type Op struct{
	LeaseID int64
}

type OpOption func(*Op)

func WithLease(leaseID int64) OpOption{
	return  func(o *Op){o.LeaseID = leaseID}
}


func (o *Op)applyOpts(ops []OpOption){
	for _,opOption := range ops{
		opOption(o)
	}
}



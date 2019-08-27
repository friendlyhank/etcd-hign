package embed

import (
	"context"
	"fmt"
)

type Kv struct{
}

func (kv *Kv) Put(ctx context.Context, key, val string, opts ...OpOption) error {
	op := &Op{}
	op.applyOpts(opts)

	kv.Do(ctx,op)
	return nil
}

func (kv *Kv)Do(ctx context.Context,op *Op){
	fmt.Println(op.LeaseID)
}
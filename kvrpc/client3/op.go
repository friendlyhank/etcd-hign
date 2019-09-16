package client3

type opType int

const (
	// A default Op has opType 0, which is invalid.
	tRange opType = iota + 1
	tPut
	tDeleteRange
)

// Op represents an Operation that kv can execute.
type Op struct {
	t opType //操作类型
}

// OpOption configures Operations like Get, Put, Delete.
type OpOption func(*Op)

func (op *Op) applyOpts(opts []OpOption) {
	for _, opt := range opts {
		opt(op)
	}
}

// OpPut returns "put" operation based on given key-value and operation options.
func OpPut(key, val string, opts ...OpOption) Op {
	ret := Op{t: tPut, key: []byte(key), val: []byte(val)}
	ret.applyOpts(opts)
	switch{

	}
	return ret
}
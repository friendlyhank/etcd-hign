package rafthttp

import (
	"io"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"
)

// messageEncoder is a encoder that can encode all kinds of messages.
// It MUST be used with a paired messageDecoder.
//v3版本的msg消息编码器
type messageEncoder struct {
	w io.Writer
}

//v3版本的msg消息解码器
// messageDecoder is a decoder that can decode all kinds of messages.
type messageDecoder struct {
	r io.Reader
}

var (
	readBytesLimit uint64 = 512 * 1024 * 1024
)

func (dec *messageDecoder) decode() (raftpb.Message, error) {
	return dec.decodeLimit(readBytesLimit)
}

func (dec *messageDecoder) decodeLimit(numBytes uint64) (raftpb.Message, error) {
	var m raftpb.Message
	return m, nil
}

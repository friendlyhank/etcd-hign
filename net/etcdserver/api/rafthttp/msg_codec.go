package rafthttp

import (
	"io"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"
)

//v3版本的msg消息解码器
// messageDecoder is a decoder that can decode all kinds of messages.
type messageDecoder struct {
	r io.Reader
}

func (dec *messageDecoder) decode() (raftpb.Message, error) {
	return dec.decodeLimit()
}

func (dec *messageDecoder) decodeLimit() (raftpb.Message, error) {
	var m raftpb.Message
	return m, nil
}

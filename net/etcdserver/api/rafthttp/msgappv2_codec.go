package rafthttp

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/friendlyhank/etcd-hign/net/pkg/pbutil"

	"github.com/friendlyhank/etcd-hign/net/raft/raftpb"

	"github.com/friendlyhank/etcd-hign/net/pkg/types"
)

const (
	msgTypeLinkHeartbeat uint8 = 0 //心跳类型消息
	msgTypeAppEntries    uint8 = 1
	msgTypeApp           uint8 = 2

	msgAppV2BufSize = 1024 * 1024
)

// msgappv2 stream sends three types of message: linkHeartbeatMessage,
// AppEntries and MsgApp. AppEntries is the MsgApp that is sent in
// replicate state in raft, whose index and term are fully predictable.
//
// Data format of linkHeartbeatMessage:
//心跳的格式为1个字节，只记录心跳的类型
// | offset | bytes | description |
// +--------+-------+-------------+
// | 0      | 1     | \x00        |
//
// Data format of AppEntries:
//多条entries消息的格式
// | offset | bytes | description |
// +--------+-------+-------------+
// | 0      | 1     | \x01        |  //1字节的消息类型
// | 1      | 8     | length of entries | //8个字节的entries消息条数
// | 9      | 8     | length of first entry |//第9自己到第16字节记录每条entry消息的数据大小
// | 17     | n1    | first entry | //第17字节开始记录每条entry消息数据
// ...
// | x      | 8     | length of k-th entry data |
// | x+8    | nk    | k-th entry data |
// | x+8+nk | 8     | commit index |
//
// Data format of MsgApp:
//单条主消息的格式
// | offset | bytes | description |
// +--------+-------+-------------+
// | 0      | 1     | \x02        | //1字节的消息类型
// | 1      | 8     | length of encoded message |//8记录消息的长度
// | 9      | n     | encoded message |  //第9字节开始记录具体的消息数据
type msgAppV2Encoder struct {
	w io.Writer

	term      uint64
	index     uint64
	buf       []byte
	uint64buf []byte
	uint8buf  []byte
}

func newMsgAppV2Encoder(w io.Writer) *msgAppV2Encoder {
	return &msgAppV2Encoder{
		w:         w,
		buf:       make([]byte, msgAppV2BufSize),
		uint64buf: make([]byte, 8),
		uint8buf:  make([]byte, 1),
	}
}

func (enc *msgAppV2Encoder) encode(m *raftpb.Message) error {
	//start := time.Now()
	switch {
	case isLinkHeartbeatMessage(m):
		enc.uint8buf[0] = msgTypeLinkHeartbeat
		if _, err := enc.w.Write(enc.uint8buf); err != nil {
			return err
		}
	case enc.index == m.Index && enc.term == m.LogTerm && m.LogTerm == m.Term: //多条消息会被整合成一条
		//将消息类型存入1字节的uint8buf
		enc.uint8buf[0] = msgTypeAppEntries
		if _, err := enc.w.Write(enc.uint8buf); err != nil {
			return err
		}
		// write length of entries
		//将entries的长度写入enc.w buf中
		binary.BigEndian.PutUint64(enc.uint64buf, uint64(len(m.Entries)))
		if _, err := enc.w.Write(enc.uint64buf); err != nil {
			return err
		}
		for i := 0; i < len(m.Entries); i++ {
			// write length of entry
			//将每个entry数据长度写入enc.w buf中
			binary.BigEndian.PutUint64(enc.uint64buf, uint64(m.Entries[i].Size()))
			if _, err := enc.w.Write(enc.uint64buf); err != nil {
				return err
			}
			//最后将序列化的entry数据写入enc.w buf中
			if n := m.Entries[i].Size(); n < msgAppV2BufSize {
				if _, err := m.Entries[i].MarshalTo(enc.buf); err != nil {
					return err
				}
				if _, err := enc.w.Write(enc.buf[:n]); err != nil {
					return err
				}
			} else {
				if _, err := enc.w.Write(pbutil.MustMarshal(&m.Entries[i])); err != nil {
					return err
				}
			}
		}
		// write commit index
		binary.BigEndian.PutUint64(enc.uint64buf, m.Commit)
		if _, err := enc.w.Write(enc.uint64buf); err != nil {
			return err
		}
	default:
		//使用大端序存储1字节的数据类型
		if err := binary.Write(enc.w, binary.BigEndian, msgTypeApp); err != nil {
			return err
		}
		// write size of message
		//使用大端序存储8个字节存储数据的大小
		if err := binary.Write(enc.w, binary.BigEndian, uint64(m.Size())); err != nil {
			return err
		}
		//write message
		if _, err := enc.w.Write(pbutil.MustMarshal(m)); err != nil {
			return err
		}

		enc.term = m.Term
		enc.index = m.Index
		if l := len(m.Entries); l > 0 {
			enc.index = m.Entries[l-1].Index
		}
	}
	return nil
}

// msgappv2 stream sends three types of message: linkHeartbeatMessage,
// AppEntries and MsgApp. AppEntries is the MsgApp that is sent in
// replicate state in raft, whose index and term are fully predictable.
//
// Data format of linkHeartbeatMessage:
// | offset | bytes | description |
// +--------+-------+-------------+
// | 0      | 1     | \x00        |
//
// Data format of AppEntries:
// | offset | bytes | description |
// +--------+-------+-------------+
// | 0      | 1     | \x01        |
// | 1      | 8     | length of entries |
// | 9      | 8     | length of first entry |
// | 17     | n1    | first entry |
// ...
// | x      | 8     | length of k-th entry data |
// | x+8    | nk    | k-th entry data |
// | x+8+nk | 8     | commit index |
//
// Data format of MsgApp:
// | offset | bytes | description |
// +--------+-------+-------------+
// | 0      | 1     | \x02        |
// | 1      | 8     | length of encoded message |
// | 9      | n     | encoded message |
//TODO Hank分析下消息的结构
//v2 msg消息解码器
type msgAppV2Decoder struct {
	r             io.Reader
	local, remote types.ID

	buf       []byte
	uint64buf []byte
	uint8buf  []byte
}

func newMsgAppV2Decoder(r io.Reader, local, remote types.ID) *msgAppV2Decoder {
	return &msgAppV2Decoder{
		r:         r,
		local:     local,
		remote:    remote,
		buf:       make([]byte, msgAppV2BufSize),
		uint64buf: make([]byte, 8),
		uint8buf:  make([]byte, 1),
	}
}

//v2版本消息的解码
func (dec *msgAppV2Decoder) decode() (raftpb.Message, error) {
	var (
		m   raftpb.Message
		typ uint8
	)
	//读取第一个字节的消息类型
	if _, err := io.ReadFull(dec.r, dec.uint8buf); err != nil {
		return m, err
	}
	typ = dec.uint8buf[0]
	switch typ {
	case msgTypeLinkHeartbeat:
		return linkHeartbeatMessage, nil
	case msgTypeAppEntries:
		m = raftpb.Message{
			Type: raftpb.MsgApp,
			From: uint64(dec.remote),
			To:   uint64(dec.local),
		}

		// decode entries
		if _, err := io.ReadFull(dec.r, dec.uint64buf); err != nil {
			return m, err
		}
		//获取entries的长度
		l := binary.BigEndian.Uint64(dec.uint64buf)
		m.Entries = make([]raftpb.Entry, int(l))
		for i := 0; i < int(l); i++ {
			if _, err := io.ReadFull(dec.r, dec.uint64buf); err != nil {
				return m, err
			}
			//获取entry的长度大小
			size := binary.BigEndian.Uint64(dec.uint64buf)
			var buf []byte
			if size < msgAppV2BufSize {
				buf = dec.buf[:size]
				if _, err := io.ReadFull(dec.r, buf); err != nil {
					return m, err
				}
			} else {
				buf = make([]byte, int(size))
				if _, err := io.ReadFull(dec.r, buf); err != nil {
					return m, err
				}
			}
			// 1 alloc
			pbutil.MustUnmarshal(&m.Entries[i], buf)
		}
		// decode commit index
		if _, err := io.ReadFull(dec.r, dec.uint64buf); err != nil {
			return m, err
		}
		m.Commit = binary.BigEndian.Uint64(dec.uint64buf)
	case msgTypeApp:
		var size uint64
		if err := binary.Read(dec.r, binary.BigEndian, &size); err != nil {
			return m, err
		}
		buf := make([]byte, int(size))
		if _, err := io.ReadFull(dec.r, buf); err != nil {
			return m, err
		}
		//解读消息体
		pbutil.MustUnmarshal(&m, buf)
	default:
		return m, fmt.Errorf("failed to parse type %d in msgappv2 stream", typ)
	}
	return m, nil
}

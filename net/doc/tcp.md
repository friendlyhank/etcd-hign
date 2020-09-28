
### tcp五个基本步骤
- 创建socket套接字
- bind地址及端口
- listen监听服务
- accept接收客户端连接
- 启动线程为客户端服务


### cmux如何去封装的


#### node
type node struct {
	propc      chan msgWithResult
	recvc      chan pb.Message
	confc      chan pb.ConfChangeV2
	confstatec chan pb.ConfState
	readyc     chan Ready
	advancec   chan struct{}
	tickc      chan struct{}
	done       chan struct{}
	stop       chan struct{}
	status     chan chan Status

	rn *RawNode
}
彻底理解结构体propc,recvc,readyc

propc应该是发送消息的结果
recvc发送消息
ready消息的准备体


在raft\node.go (n *node)run()去设置ready,让节点进入准备状态
当节点进入准备状态就会在etcdserver\raft.go 
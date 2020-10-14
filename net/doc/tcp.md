## node
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

node作为消息的最上层应用
node会有对应的raft协议，然后去对应的发送消息
只有在节点readyc状态的情况下，才能去发送消息


在raft\node.go (n *node)run()去设置ready,让节点进入准备状态
当节点进入准备状态就会在etcdserver\raft.go  (r *raftNode) start(rh *raftReadyHandler)不断去发送消息

## Transport
transport是网络的核心组件，所有我网络入口都会先进入transport
transport作为最中间的一层

然后transport对应三大Handle
pipelinehandle
streamhandle
snaphandle

每个handle会走到对象的peer节点

## peer 
### stream 流发送少量消息 用完不会立刻关闭
peer对应的是多个流
    msgAppV2Writer v2的写入流
    writer 最新版本的写入流
    msgAppV2Reader v2的读取流
    msgAppReader 最新版本的读取流

### pipeline 管道去发送多个消息 用完马上关闭

## 核心组件库transport
    最核心的网络组件库,也是最底层的组件库

### cmux如何去封装的

### tcp五个基本步骤
- 创建socket套接字
- bind地址及端口
- listen监听服务
- accept接收客户端连接
- 启动线程为客户端服务

## 主要过程
三个节点A,B,C
当nodeA处于准备准备状态,则会发送消息给B,C投票
        from:A
        To:B
        
        form:A
        To:C

在底层网络启动的时候,peer.msgAppReader首先回去尝试连接,会在url中带上节点id信息
        from:url的节点id
        to:B,C

在底层网络启动的时候，peer.writer网络处于接收状态
     from:url的节点id B,C
     to:当前节点 A

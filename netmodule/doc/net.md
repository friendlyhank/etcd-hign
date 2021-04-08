最新参照的版本etcd-3.4.13

## node
type node struct {
	propc      chan msgWithResult
	recvc      chan pb.Message
	readyc     chan Ready //节点的就绪状态,只有就绪时候才能去发送消息
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

### 初始化和启动
new一个http.RoundTripper来完成一个http事务,就是客户端发送一个Http Request到服务端,
服务端处理完请求之后，返回相应的HTTP Response的过程
t.streamRt, err = newStreamRoundTripper(t.TLSInfo, t.DialTimeout)
t.pipelineRt, err = NewRoundTripper(t.TLSInfo, t.DialTimeout)

Transport RoundTripper可以处理不同的协议

然后transport对应三大
Handle客户端的处理逻辑和相应逻辑
pipelinehandle
streamhandle
snaphandle

## peer 
stream消息通道维护HTTP长连接，主要负责传输数据小、发送比较频繁的消息，例如MsgApp消息、MsgHeartbeat消息、MsgVote消息等;
而Pipleline消息通道在传输数据完成之后会立即关闭连接，主要负责数据量大、发送频率较低的消息,例如:MsgSnap消息等。

### stream
peer对应的是多个流
    msgAppV2Writer v2的写入流
    writer 最新版本的写入流
    
在peer.start()方法中通过调用startStreamWriter()方法初始化并启动streamWriter实例，其中还启动了一个后台goroutine来执行
streamWriter.run()方法中，主要完成了下面三件事：
(1)当其他节点与当前节点创建连接时,该连接实例会写入对应peer.writer.conn通道，在streamWriter.run()方法中通过该通道获取该连接之后进行绑定
,之后才能开始后续的消息发送。
(2)定时发送心跳
    
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


问题：
具体这些消息会进入那些通道


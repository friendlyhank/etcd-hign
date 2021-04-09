package raft

import pb "github.com/friendlyhank/etcd-hign/raftmodule/raft/raftpb"

// Bootstrap initializes the RawNode for first use by appending configuration
// changes for the supplied peers. This method returns an error if the Storage
// is nonempty.
//
// It is recommended that instead of calling this method, applications bootstrap
// their state manually by setting up a Storage that has a first index > 1 and
// which stores the desired ConfState as its InitialState.
func (rn *RawNode) Bootstrap(peers []Peer) error {

	// TODO(tbg): remove StartNode and give the application the right tools to
	// bootstrap the initial membership in a cleaner way.
	//设置成为跟随者的方法
	rn.raft.becomeFollower(1, None)

	//申请修改raft的配置信息,事件为新增投票的节点
	for _, peer := range peers {
		rn.raft.applyConfChange(pb.ConfChange{NodeID: peer.ID, Type: pb.ConfChangeAddNode}.AsV2())
	}
	return nil
}

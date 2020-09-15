package election

// raft 协议
type raft struct{
	id uint64

	Term uint64
	Vote uint64
}

// raftNode -
type raftNode struct{

}
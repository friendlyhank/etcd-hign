package raft

// Bootstrap initializes the RawNode for first use by appending configuration
// changes for the supplied peers. This method returns an error if the Storage
// is nonempty.
//
// It is recommended that instead of calling this method, applications bootstrap
// their state manually by setting up a Storage that has a first index > 1 and
// which stores the desired ConfState as its InitialState.
func (rn *RawNode) Bootstrap() error {

	// TODO(tbg): remove StartNode and give the application the right tools to
	// bootstrap the initial membership in a cleaner way.
	//设置成为跟随者的方法
	rn.raft.becomeFollower(1, None)

	return nil
}

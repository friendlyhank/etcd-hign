package membership

type RaftCluster struct {
}

func NewClusterFromURLsMap() (*RaftCluster, error) {
	c := NewCluster()
	return c, nil
}

func NewCluster() *RaftCluster {
	return &RaftCluster{}
}

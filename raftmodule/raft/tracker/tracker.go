package tracker

//raft统计票数相关

// Config reflects the configuration tracked in a ProgressTracker.
type Config struct {
}

// ProgressTracker tracks the currently active configuration and the information
// known about the nodes and learners in it. In particular, it tracks the match
// index for each peer which in turn allows reasoning about the committed index.
//程序投票统计
type ProgressTracker struct {
	Config
	Progress ProgressMap

	Votes map[uint64]bool
}

// MakeProgressTracker initializes a ProgressTracker.
func MakeProgressTracker(maxInflight int) ProgressTracker {
	p := ProgressTracker{
		Config:   Config{},
		Votes:    map[uint64]bool{},
		Progress: map[uint64]*Progress{},
	}
	return p
}

// ProgressMap is a map of *Progress.
type ProgressMap map[uint64]*Progress

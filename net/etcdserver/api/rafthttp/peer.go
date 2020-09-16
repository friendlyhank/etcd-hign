package rafthttp

import "time"

const(
	ConnReadTimeout  = 5 * time.Second
	ConnWriteTimeout = 5 * time.Second
)

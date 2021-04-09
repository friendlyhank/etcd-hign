package raftpb

// AsV2 returns a V2 configuration change carrying out the same operation.
func (c ConfChange) AsV2() ConfChangeV2 {
	return ConfChangeV2{
		Changes: []ConfChangeSingle{{
			Type:   c.Type,
			NodeID: c.NodeID,
		}},
		Context: c.Context,
	}
}

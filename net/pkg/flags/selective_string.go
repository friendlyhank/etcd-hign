package flags

import (
	"errors"
	"sort"
)

// SelectiveStringValue implements the flag.Value interface.
type SelectiveStringValue struct {
	v      string
	valids map[string]struct{}
}

// Set verifies the argument to be a valid member of the allowed values
// before setting the underlying flag value.
func (ss *SelectiveStringValue) Set(s string) error {
	if _, ok := ss.valids[s]; ok {
		ss.v = s
		return nil
	}
	return errors.New("invalid value")
}

// String returns the set value (if any) of the SelectiveStringValue
func (ss *SelectiveStringValue) String() string {
	return ss.v
}

// Valids returns the list of valid strings.
func (ss *SelectiveStringValue) Valids() []string {
	s := make([]string, 0, len(ss.valids))
	for k := range ss.valids {
		s = append(s, k)
	}
	sort.Strings(s)
	return s
}

// NewSelectiveStringValue creates a new string flag
// for which any one of the given strings is a valid value,
// and any other value is an error.
//
// valids[0] will be default value. Caller must be sure
// len(valids) != 0 or it will panic.
func NewSelectiveStringValue(valids ...string) *SelectiveStringValue {
	vm := make(map[string]struct{})
	for _, v := range valids {
		vm[v] = struct{}{}
	}
	return &SelectiveStringValue{valids: vm, v: valids[0]}
}

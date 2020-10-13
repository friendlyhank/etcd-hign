package flags

import (
	"flag"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"go.etcd.io/etcd/pkg/types"
)

type UniqueURLs struct {
	Values  map[string]struct{}
	uss     []url.URL
	Allowed map[string]struct{}
}

// Set parses a command line set of URLs formatted like:
// http://127.0.0.1:2380,http://10.1.1.2:80
// Implements "flag.Value" interface.
func (us *UniqueURLs) Set(s string) error {
	if _, ok := us.Values[s]; ok {
		return nil
	}
	if _, ok := us.Allowed[s]; ok {
		us.Values[s] = struct{}{}
		return nil
	}
	ss, err := types.NewURLs(strings.Split(s, ","))
	if err != nil {
		return err
	}
	us.Values = make(map[string]struct{})
	us.uss = make([]url.URL, 0)
	for _, v := range ss {
		us.Values[v.String()] = struct{}{}
		us.uss = append(us.uss, v)
	}
	return nil
}

// String implements "flag.Value" interface.
func (us *UniqueURLs) String() string {
	all := make([]string, 0, len(us.Values))
	for u := range us.Values {
		all = append(all, u)
	}
	sort.Strings(all)
	return strings.Join(all, ",")
}

// NewUniqueURLsWithExceptions implements "url.URL" slice as flag.Value interface.
// Given value is to be separated by comma.
func NewUniqueURLsWithExceptions(s string, exceptions ...string) *UniqueURLs {
	us := &UniqueURLs{Values: make(map[string]struct{}), Allowed: make(map[string]struct{})}
	for _, v := range exceptions {
		us.Allowed[v] = struct{}{}
	}
	if s == "" {
		return us
	}
	if err := us.Set(s); err != nil {
		panic(fmt.Sprintf("new UniqueURLs should never fail: %v", err))
	}
	return us
}

// UniqueURLsFromFlag returns a slice from urls got from the flag.
func UniqueURLsFromFlag(fs *flag.FlagSet, urlsFlagName string) []url.URL {
	return (*fs.Lookup(urlsFlagName).Value.(*UniqueURLs)).uss
}

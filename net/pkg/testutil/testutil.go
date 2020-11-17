package testutil

import (
	"net/url"
	"testing"
)

func MustNewURLs(t *testing.T, urls []string) []url.URL {
	if urls == nil {
		return nil
	}
	var us []url.URL
	for _, url := range urls {
		u := MustNewURL(t, url)
		us = append(us, *u)
	}
	return us
}

func MustNewURL(t *testing.T, s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		t.Fatalf("pase %v error: %v", s, err)
	}
	return u
}

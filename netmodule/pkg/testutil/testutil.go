package testutil

import (
	"net/url"
	"testing"
	"time"
)

// WaitSchedule briefly sleeps in order to invoke the go scheduler.
// TODO: improve this when we are able to know the schedule or status of target go-routine.
func WaitSchedule() {
	time.Sleep(10 * time.Millisecond)
}

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

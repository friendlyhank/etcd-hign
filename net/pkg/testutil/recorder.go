package testutil

import (
	"errors"
	"fmt"
	"time"
)

type Action struct {
	Name   string
	Params []interface{}
}

type Recorder interface {
	// Record publishes an Action (e.g., function call) which will
	// be reflected by Wait() or Chan()
	Record(a Action)
	// Wait waits until at least n Actions are available or returns with error
	Wait(n int) ([]Action, error)
	// Action returns immediately available Actions
	Action() []Action
	// Chan returns the channel for actions published by Record
	Chan() <-chan Action
}

// RecorderStream writes all Actions to an unbuffered channel
type recorderStream struct {
	ch          chan Action
	waitTimeout time.Duration
}

func NewRecorderStream() Recorder {
	return NewRecorderStreamWithWaitTimout(time.Duration(5 * time.Second))
}

func NewRecorderStreamWithWaitTimout(waitTimeout time.Duration) Recorder {
	return &recorderStream{ch: make(chan Action), waitTimeout: waitTimeout}
}

func (r *recorderStream) Record(a Action) {
	r.ch <- a
}

func (r *recorderStream) Action() (acts []Action) {
	for {
		select {
		case act := <-r.ch:
			acts = append(acts, act)
		default:
			return acts
		}
	}
}

func (r *recorderStream) Chan() <-chan Action {
	return r.ch
}

func (r *recorderStream) Wait(n int) ([]Action, error) {
	acts := make([]Action, n)
	timeoutC := time.After(r.waitTimeout)
	for i := 0; i < n; i++ {
		select {
		case acts[i] = <-r.ch:
		case <-timeoutC:
			acts = acts[:i]
			return acts, newLenErr(n, i)
		}
	}
	// extra wait to catch any Action spew
	select {
	case act := <-r.ch:
		acts = append(acts, act)
	case <-time.After(10 * time.Millisecond):
	}
	return acts, nil
}

func newLenErr(expected int, actual int) error {
	s := fmt.Sprintf("len(actions) = %d, expected >= %d", actual, expected)
	return errors.New(s)
}
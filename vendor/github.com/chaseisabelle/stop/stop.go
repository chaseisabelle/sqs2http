package stop

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var state struct {
	sync.Mutex
	interrupted bool
	killed bool
	terminated bool
}

func Listen() {
	state = struct {
		sync.Mutex
		interrupted bool
		killed bool
		terminated bool
	}{}

	c := make(chan os.Signal)

	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	go func() {
		s := <-c

		state.Lock()

		defer state.Unlock()

		switch s {
		case syscall.SIGINT:
			state.interrupted = true
		case syscall.SIGKILL:
			state.killed = true
		case syscall.SIGTERM:
			state.terminated = true
		}
	}()
}

func Interrupted() bool {
	state.Lock()

	defer state.Unlock()

	return state.interrupted
}

func Killed() bool {
	state.Lock()

	defer state.Unlock()

	return state.killed
}

func Terminated() bool {
	state.Lock()

	defer state.Unlock()

	return state.terminated
}

func Stopped() bool {
	state.Lock()

	defer state.Unlock()

	return state.interrupted || state.terminated || state.killed
}

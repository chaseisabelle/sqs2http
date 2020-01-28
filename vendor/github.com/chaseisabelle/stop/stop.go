package stop

import (
	"os"
	"os/signal"
	"sync"
)

type state struct {
	sync.Mutex
	interrupted bool
}

var s state

func Listen() {
	c := make(chan os.Signal)

	signal.Notify(c, os.Interrupt)

	go func() {
		<-c

		s.Lock()

		defer s.Unlock()

		s.interrupted = true
	}()
}

func Interrupted() bool {
	s.Lock()

	defer s.Unlock()

	return s.interrupted
}

package backoff

import (
	"errors"
	"sync"
	"time"
)

type Backoff struct {
	sync.Mutex
	initial uint64
	current uint64
	max     uint64
	algo    func(uint64) uint64
}

func New(initial uint64, current uint64, max uint64, algo func(uint64) uint64) (*Backoff, error) {
	if initial > max {
		return nil, errors.New("initial cant be greater than max")
	}

	if current > max {
		return nil, errors.New("current cant be greater than max")
	}

	if algo == nil {
		algo = func(current uint64) uint64 {
			return current
		}
	}

	return &Backoff{
		initial: initial,
		current: current,
		max:     max,
		algo:    algo,
	}, nil
}

func (bo *Backoff) Backoff() {
	bo.Lock()

	defer bo.Unlock()

	time.Sleep(time.Duration(bo.current) * time.Millisecond)

	bo.current = bo.algo(bo.current)

	if bo.current > bo.max {
		bo.current = bo.max
	}
}

func (bo *Backoff) Reset() {
	bo.Lock()

	defer bo.Unlock()

	bo.current = bo.initial
}

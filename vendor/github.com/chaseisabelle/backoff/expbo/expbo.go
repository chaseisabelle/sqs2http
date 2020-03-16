package expbo

import (
	"github.com/chaseisabelle/backoff"
	"math"
)

func New(initial uint64, max uint64, exp uint64) (*backoff.Backoff, error) {
	return backoff.New(initial, initial, max, func(current uint64) uint64 {
		return uint64(math.Pow(float64(current), float64(exp)))
	})
}
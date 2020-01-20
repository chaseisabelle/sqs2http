package flagz

import (
	"strings"
)

type Flagz []string

func (s *Flagz) Array() []string {
	return *s
}

func (s *Flagz) String() string {
	return strings.Join(s.Array(), ", ")
}

func (s *Flagz) Set(str string) error {
	*s = append(*s, str)

	return nil
}



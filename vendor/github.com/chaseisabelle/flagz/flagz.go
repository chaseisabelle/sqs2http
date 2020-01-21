package flagz

import (
	"strconv"
	"strings"
)

type Flagz []string

func (s *Flagz) String() string {
	return strings.Join(s.Array(), ", ")
}

func (s *Flagz) Set(str string) error {
	*s = append(*s, str)

	return nil
}

func (s *Flagz) Array() []string {
	return *s
}

func (s *Flagz) Stringz() []string {
	return s.Array()
}

func (s *Flagz) Intz() ([]int, error) {
	strs := s.Stringz()
	ints := make([]int, len(strs))

	for index, str := range strs {
		int, err := strconv.Atoi(str)

		if err != nil {
			return ints, err
		}

		ints[index] = int
	}

	return ints, nil
}

func (s *Flagz) Boolz() ([]bool, error) {
	strs := s.Stringz()
	boolz := make([]bool, len(strs))

	for index, str := range strs {
		b, err := strconv.ParseBool(str)

		if err != nil {
			return boolz, err
		}

		boolz[index] = b
	}

	return boolz, nil
}

func (s *Flagz) Floatz() ([]float64, error) {
	strs := s.Stringz()
	floatz := make([]float64, len(strs))

	for index, str := range strs {
		f, err := strconv.ParseFloat(str, 64)

		if err != nil {
			return floatz, err
		}

		floatz[index] = f
	}

	return floatz, nil
}



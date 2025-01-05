package main

import "errors"

type Stack struct {
	data []int32
}

func (s *Stack) Push(value int32) {
	s.data = append(s.data, value)
}

func (s *Stack) Pop() (int32, error) {
	if len(s.data) == 0 {
		return 0, errors.New("stack underflow")
	}
	value := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return value, nil
}

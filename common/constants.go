package common

import "errors"

var (
	ErrOverflow = errors.New("data structure overflow")
	ErrEmpty    = errors.New("data structure is empty")
)

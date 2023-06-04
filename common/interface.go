package common

import "errors"

var (
	ErrOverflow = errors.New("data overflow")
)

type DataStructureFactory[Value any] func(capacity int) DataStructure[Value]

type DataStructure[Value any] interface {
	Push(Value) bool
	Pop() (Value, bool)
	Size() int
	Empty() bool
	Full() bool
}

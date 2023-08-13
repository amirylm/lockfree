package pool

import (
	"context"
	"crypto/sha256"
	"errors"
	"hash"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPoolWrapper(t *testing.T) {
	sha256Hash := PoolWrapper(sha256.New, func(h hash.Hash, data []byte) [32]byte {
		var b [32]byte
		h.Reset()
		// #nosec G104
		_, err := h.Write(data)
		if err != nil {
			return b
		}
		h.Sum(b[:0])

		return b
	})

	done := make(chan bool, 1)
	go func() {
		defer func() {
			done <- true
		}()
		input, expected := bytesInput()
		require.Equal(t, expected, sha256Hash(input))
	}()
	input, expected := bytesInput()
	require.Equal(t, expected, sha256Hash(input))

	<-done
}

// TODO: move to benchmarks
func TestPoolWrapperLoad(t *testing.T) {
	sha256Hash := PoolWrapper(sha256.New, func(h hash.Hash, data []byte) [32]byte {
		var b [32]byte
		h.Reset()
		// #nosec G104
		_, err := h.Write(data)
		if err != nil {
			return b
		}
		h.Sum(b[:0])

		return b
	})

	n := 1000
	workers := make(chan bool)

	for n > 0 {
		n--
		go func() {
			defer func() {
				workers <- true
			}()
			input, expected := bytesInput1024()
			require.Equal(t, expected, sha256Hash(input))
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		case <-workers:
			continue
		}
	}
}

func TestPoolWrapperWithErr(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(h hash.Hash, data []byte) ([32]byte, error)
		errored bool
	}{
		{
			"no error",
			func(h hash.Hash, data []byte) ([32]byte, error) {
				var b [32]byte
				h.Reset()
				// #nosec G104
				_, err := h.Write(data)
				if err != nil {
					return b, err
				}
				h.Sum(b[:0])

				return b, nil
			},
			false,
		},
		{
			"with error",
			func(h hash.Hash, data []byte) ([32]byte, error) {
				var b [32]byte
				return b, errors.New("test-error")
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fn := PoolWrapperWithErr(sha256.New, tc.fn)

			check := func(errored bool) {
				input, expected := bytesInput()
				h, err := fn(input)
				if errored {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.Equal(t, expected, h)
				}
			}
			done := make(chan bool, 1)
			go func(errored bool) {
				defer func() {
					done <- true
				}()
				check(errored)
			}(tc.errored)

			check(tc.errored)

			<-done
		})
	}
}

func TestPoolWrapperErr(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(string, []byte) error
		errored bool
	}{
		{
			"no error",
			func(s string, data []byte) error {
				return nil
			},
			false,
		},
		{
			"with error",
			func(s string, data []byte) error {
				return errors.New("test-error")
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fn := PoolWrapperErr(func() string {
				return "test"
			}, tc.fn)

			check := func(errored bool) {
				input, _ := bytesInput()
				err := fn(input)
				if errored {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}

			done := make(chan bool, 1)
			go func(errored bool) {
				defer func() {
					done <- true
				}()
				check(tc.errored)
			}(tc.errored)

			check(tc.errored)

			<-done
		})
	}
}

func bytesInput() ([]byte, [32]byte) {
	return []byte("test-data"), [32]uint8{0xa1, 0x86, 0x0, 0x4, 0x22, 0xfe, 0xab, 0x85, 0x73, 0x29, 0xc6, 0x84, 0xe9, 0xfe, 0x91, 0x41, 0x2b, 0x1a, 0x5d, 0xb0, 0x84, 0x10, 0xb, 0x37, 0xa9, 0x8c, 0xfc, 0x95, 0xb6, 0x2a, 0xa8, 0x67}
}

func bytesInput1024() ([]byte, [32]byte) {
	size := 1024
	input := make([]byte, size)
	str := "test-data"
	for i := 0; i < size; i++ {
		input = append(input, byte(str[i%len(str)]))
	}
	return input, [32]uint8{0x6d, 0x46, 0xcf, 0x6e, 0x4, 0x74, 0x9c, 0xd5, 0x2e, 0x3, 0x68, 0xab, 0x77, 0x14, 0xd, 0xd3, 0x9a, 0xb5, 0x6d, 0x79, 0x3c, 0x0, 0xc3, 0xa5, 0x83, 0xba, 0x92, 0xaf, 0x32, 0x15, 0x63, 0x1f}
}

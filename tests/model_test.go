//go:build model_test

package fibonacci

import (
	"slices"
	"sync"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

// TODO: в ридми начало с 1

func TestGoldenSequence(t *testing.T) {
	t.Parallel()

	generator := NewGenerator()
	fibSequence := make([]uint64, 0, 9)

	expected := [9]uint64{0, 1, 1, 2, 3, 5, 8, 13, 21}

	for range 9 {
		fibSequence = append(fibSequence, generator.Next())
	}

	require.Equal(t, expected[:], fibSequence)
}

func TestOverflow(t *testing.T) {
	t.Parallel()
	generator := NewGenerator()

	var prev uint64
	for range maxFibonacciNumber {
		cur := generator.Next()
		require.LessOrEqual(t, prev, cur)

		prev = cur
	}
}

func TestGolden(t *testing.T) {
	t.Parallel()

	// F_93 = 7540113804746346429
	const expected uint64 = 7540113804746346429

	generator := NewGenerator()
	var cur uint64

	for range maxFibonacciNumber {
		cur = generator.Next()
	}

	require.EqualValues(t, expected, cur)
}

func TestPanicOnOverflow(t *testing.T) {
	t.Parallel()
	generator := NewGenerator()

	var prev uint64
	for range maxFibonacciNumber {
		cur := generator.Next()
		require.LessOrEqual(t, prev, cur)

		prev = cur
	}

	func() {
		defer func() {
			err := recover()
			require.NotNil(t, err, "expected panic on overflow")

			vErr, ok := err.(error)
			require.True(t, ok, "expected panic with error on overflow")

			require.ErrorContains(t, vErr, "overflow", "expected verbose message on overflow")
		}()

		generator.Next()
	}()
}

func TestFibonacciInvariant(t *testing.T) {
	t.Parallel()

	workers := 10
	generator := NewGenerator()
	const iters = 9

	wg := new(sync.WaitGroup)
	workerValues := make([][]uint64, workers)
	for workerID := range workers {
		workerValues[workerID] = make([]uint64, 0, iters)
	}

	for workerID := range workers {
		wg.Go(func() {
			for range iters {
				value := generator.Next()
				workerValues[workerID] = append(workerValues[workerID], value)
			}
		})
	}

	wg.Wait()

	allValues := make([]uint64, 0, iters*workers)
	for _, values := range workerValues {
		require.Truef(t, slices.IsSorted(values), "expected increasing sequence of values, got: %v", values)
		allValues = append(allValues, values...)
	}

	slices.Sort(allValues)
	require.True(t, isFibonacci(t, allValues), "expected fibonacci sequence")
}

func TestUnlockPanicOnOverflow(t *testing.T) {
	t.Parallel()
	generator := NewGenerator()

	var prev uint64
	for range maxFibonacciNumber {
		cur := generator.Next()
		require.LessOrEqual(t, prev, cur)

		prev = cur
	}

	wg := new(sync.WaitGroup)
	for range 100 {
		wg.Go(func() {
			func() {
				defer func() {
					err := recover()
					require.NotNil(t, err, "expected panic on overflow")

					vErr, ok := err.(error)
					require.True(t, ok, "expected panic with error on overflow")

					require.ErrorContains(t, vErr, "overflow", "expected verbose message on overflow")
				}()

				generator.Next()
			}()
		})
	}

	wg.Wait()
}

func TestInternalSize(t *testing.T) {
	require.Equal(t, unsafe.Sizeof(int64(0))*3, unsafe.Sizeof(generatorImpl{}))
}

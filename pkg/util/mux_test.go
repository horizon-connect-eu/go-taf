package util_test

import (
	"testing"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/util"
)

func TestMuxMany(t *testing.T) {
	for _, n := range []int{1, 2, 3, 4, 5, 10, 20, 100} {
		channels := make([]chan int, 0, n)

		for range n {
			channels = append(channels, make(chan int, 10))
		}
		out := make(chan int, 10)

		util.MuxMany(channels, out)

		for i := range n {
			go func() {
				for range n {
					channels[i] <- i
				}
				close(channels[i])
			}()
		}

		nrec := 0
		for _ = range out {
			nrec++
		}

		if nrec != n*n {
			t.Errorf("unexpected number of received elements in out channel %d instead of %d (n=%d)\n", nrec, n*n, n)
		}
	}
}

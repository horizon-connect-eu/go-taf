package util_test

import (
	"testing"

	"github.com/vs-uulm/go-taf/pkg/util"
)

func TestMuxMany(t *testing.T) {
	testcases := map[string]struct {
		NChannels       int
		NMsgsPerChannel int
	}{
		"one message, one channel":     {NChannels: 1, NMsgsPerChannel: 1},
		"one message, two channels":    {NChannels: 2, NMsgsPerChannel: 1},
		"one message, 100 channels":    {NChannels: 100, NMsgsPerChannel: 1},
		"two messages, one channel":    {NChannels: 1, NMsgsPerChannel: 2},
		"ten messages, one channel":    {NChannels: 1, NMsgsPerChannel: 10},
		"100 messages, one channel":    {NChannels: 1, NMsgsPerChannel: 100},
		"two messages, two channels":   {NChannels: 2, NMsgsPerChannel: 2},
		"100 messages, 100 channels":   {NChannels: 100, NMsgsPerChannel: 100},
		"100 messages,  1000 channels": {NChannels: 1000, NMsgsPerChannel: 100},
		"1000 messages, 100 channel":   {NChannels: 100, NMsgsPerChannel: 1000},
	}

	for name, cs := range testcases {

		t.Run(name, func(t *testing.T) {
			channels := make([]chan int, 0, cs.NChannels)

			for range cs.NChannels {
				channels = append(channels, make(chan int, cs.NMsgsPerChannel))
			}
			out := make(chan int, cs.NMsgsPerChannel)

			util.Mux(out, channels...)

			for i := range cs.NChannels {
				go func() {
					for range cs.NMsgsPerChannel {
						channels[i] <- i + 1
					}
					close(channels[i])
				}()
			}

			nrec := 0
			for rec := range out {
				nrec++
				if rec == 0 {
					t.Errorf("channel closed prematurely")
				}
			}

			nExpected := cs.NMsgsPerChannel * cs.NChannels
			if nrec != nExpected {
				t.Errorf("unexpected number of received elements in out channel %d instead of %d\n", nrec, nExpected)
			}
			if x := <-out; x != 0 {
				t.Errorf("channel expected to be closed, but open")
			}
		})

	}
}

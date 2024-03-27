package trustassessment

import (
	"testing"

	"github.com/vs-uulm/go-taf/pkg/message"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/instance"
)

func generateStates(nkeys int, nentries int) State {
	state := State{}
	for i := range nkeys {
		state[i] = instance.NewTrustModelInstance(i, "test")
	}
	return state
}

func BenchmarkUpdateWorkerState(b *testing.B) {
	cases := map[string]struct {
		nkeys    int
		nentries int
	}{
		"small state": {nkeys: 1, nentries: 2},
		"large state": {nkeys: 10_000, nentries: 100_000},
	}

	b.ResetTimer()
	for benchName, data := range cases {
		b.StopTimer()
		state := generateStates(data.nkeys, data.nentries)
		tmt := TMTs{}
		tmt["A"] = 10
		tmt["B"] = 10

		msg := message.New(0, 1, "TSM", "A")

		b.StartTimer()
		b.Run(benchName, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				updateWorkerState(state, tmt, msg)
			}
		})
	}
}

func BenchmarkUpdateWorkerResults(b *testing.B) {
	cases := map[string]struct {
		nkeys    int
		nentries int
	}{
		"small state": {nkeys: 1, nentries: 2},
		"large state": {nkeys: 10_000, nentries: 100_000},
	}

	b.ResetTimer()
	for benchName, data := range cases {
		b.StopTimer()
		state := generateStates(data.nkeys, data.nentries)
		results := Results{}
		tmt := TMTs{}

		b.StartTimer()
		b.Run(benchName, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				func(a, b, c, d any) {}(results, state, tmt, 0) // TODO remove placeholder func
			}
		})
	}
}

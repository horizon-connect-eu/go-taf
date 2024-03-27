package add

import (
	"github.com/vs-uulm/go-taf/pkg/trustassessment"
)

func init() {
	trustassessment.RegisterUpdateResultFunc("add", UpdateWorkerResultsAdd)
}

// Gets the slice stored in `states` under the key `id`, computes its sum,
// and inserts this sum into `results` at key `id`.
func UpdateWorkerResultsAdd(results trustassessment.Results, states trustassessment.State, tmts trustassessment.TMTs, id int) {
	/*
		sum := 0
		for _, x := range states[id] {
			sum += x
		}
		results[id] = sum

	*/
}

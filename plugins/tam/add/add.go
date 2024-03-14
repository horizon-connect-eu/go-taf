package main

import "gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tam"

func RegisterTam(tamInst tam.TamBuilder) {
	tamInst.SetUpdateResults(UpdateWorkerResultsAdd)
}

// Gets the slice stored in `states` under the key `id`, computes its sum,
// and inserts this sum into `results` at key `id`.
func UpdateWorkerResultsAdd(results tam.Results, states tam.State, tmts tam.TMTs, id int) {
	sum := 0
	for _, x := range states[id] {
		sum += x
	}
	results[id] = sum
	//log.Printf("Current sum for ID %d: %d\n", id, sum)
}

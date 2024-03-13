package main

import "gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tam"

// Gets the slice stored in `states` under the key `id`, computes its product,
// and inserts this sum into `results` at key `id`.
func UpdateWorkerResultsMult(results tam.Results, states tam.State, tmts tam.TMTs, id int) {
	prod := 1
	for _, x := range states[id] {
		prod *= x
	}
	results[id] = prod
	//log.Printf("Current sum for ID %d: %d\n", id, sum)
}

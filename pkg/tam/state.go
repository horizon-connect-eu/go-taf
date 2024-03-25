package tam

import "gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"

// State is the collection of a number of Trust Model Instances.
// It holds all TMIs owned by one worker goroutine.
// It does not include the Results.
type State = map[int][]int

// A function creating a new State.
type StateFactory = func() State

// A function that updates the State of a worker given an incoming message.
type StateUpdater = func(State, TMTs, message.Message)

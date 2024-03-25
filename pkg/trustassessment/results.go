package trustassessment

type Results = map[int]int

type ResultsFactory = func() Results

type ResultsUpdater func(Results, State, TMTs, int)

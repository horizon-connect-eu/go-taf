package tlee

import (
	"errors"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"log/slog"
)

/*
TLEE represents an internal TLEE as part of the TAF that can be used for debugging purposes, independent of the actual
TLEE implementation.
*/
type TLEE struct {
	logger        *slog.Logger
	debuggingMode bool
}

type CurrentEntry struct {
	source      string
	destination string
	opinions    []subjectivelogic.Opinion
}

func SpawnNewTLEE(logger *slog.Logger, filePath string, debuggingMode bool) *TLEE {
	return &TLEE{
		logger:        logger,
		debuggingMode: debuggingMode,
	}

}

func (t *TLEE) RunTLEE(trustmodelID string, version int, fingerprint uint32, structure trustmodelstructure.TrustGraphStructure, values map[string][]trustmodelstructure.TrustRelationship) (map[string]subjectivelogic.QueryableOpinion, error) {
	results := make(map[string]subjectivelogic.QueryableOpinion)

	var ff func(opinion1 *subjectivelogic.Opinion, opinion2 *subjectivelogic.Opinion) (subjectivelogic.Opinion, error)

	switch structure.Operator() {
	case trustmodelstructure.AveragingFusion:
		ff = subjectivelogic.AveragingFusion
	case trustmodelstructure.ConstraintFusion:
		ff = subjectivelogic.ConstraintFusion
	case trustmodelstructure.CumulativeFusion:
		ff = subjectivelogic.CumulativeFusion
	case trustmodelstructure.WeightedFusion:
		ff = subjectivelogic.WeightedFusion
	case trustmodelstructure.NoFusion:
		for scope, relationships := range values {
			if len(relationships) != 1 {
				return nil, errors.New("No Fusion Operator Provided, although required")
			}
			results[scope] = relationships[0].Opinion()
		}
	default:
		return nil, errors.New("Unsupported Fusion Operator") // + structure.Operator())
	}

	for scope, relationships := range values {
		indegrees := make(map[string]int)
		current := make(map[string]CurrentEntry)

		numVertices := 0
		nodes := make(map[string]int)

		for _, relationship := range relationships {
			indegrees[relationship.Destination()]++

			key := relationship.Source() + ":" + relationship.Destination()

			opinion, err := subjectivelogic.NewOpinion(
				relationship.Opinion().Belief(),
				relationship.Opinion().Disbelief(),
				relationship.Opinion().Uncertainty(),
				relationship.Opinion().BaseRate(),
			)
			if err != nil {
				return nil, errors.New("subjective logic error" + err.Error())
			}

			entry, exists := current[key]
			if !exists {
				current[key] = CurrentEntry{
					source:      relationship.Source(),
					destination: relationship.Destination(),
					opinions:    []subjectivelogic.Opinion{opinion},
				}
			} else {
				entry.opinions = append(entry.opinions, opinion)
				current[key] = entry
			}

			if _, exists := nodes[relationship.Source()]; !exists {
				nodes[relationship.Source()] = numVertices
				numVertices += 1
			}
			if _, exists := nodes[relationship.Destination()]; !exists {
				nodes[relationship.Destination()] = numVertices
				numVertices += 1
			}
		}

	L:
		for {
			// Initializing result matrix and filling it up with same values as given graph
			reverseNodes := make([]string, numVertices)
			for node, i := range nodes {
				reverseNodes[i] = node
			}

			prev := make([][]int, numVertices)
			dist := make([][]int, numVertices)
			for i := 0; i < numVertices; i++ {
				prev[i] = make([]int, numVertices)
				dist[i] = make([]int, numVertices)
				for j := 0; j < numVertices; j++ {
					prev[i][j] = -1
					dist[i][j] = 999
				}
			}

			for key, entry := range current {
				prev[nodes[entry.source]][nodes[entry.destination]] = nodes[entry.source]
				dist[nodes[entry.source]][nodes[entry.destination]] = -1

				if len(entry.opinions) > 1 {
					// fuse
					var prev subjectivelogic.Opinion
					for i, v := range entry.opinions {
						if i == 0 {
							prev = v
						} else {
							fused, err := ff(&prev, &v)
							if err != nil {
								return nil, errors.New("cannot fuse opinions" + err.Error())
							}

							prev = fused
						}
					}

					// store fused opinion
					entry.opinions = []subjectivelogic.Opinion{prev}
					current[key] = entry
				}
			}

			if len(current) == 1 {
				for _, u := range current {
					results[scope] = &u.opinions[0]
				}

				break
			}

			var lowestSource int
			var lowestDestination int

			// Running over the result matrix and following the algorithm
			for k := 0; k < numVertices; k++ {
				for i := 0; i < numVertices; i++ {
					for j := 0; j < numVertices; j++ {
						// If there is a less costly path from i to j node, remembering it
						if dist[i][j] > dist[i][k]+dist[k][j] {
							dist[i][j] = dist[i][k] + dist[k][j]
							prev[i][j] = prev[k][j]

							if dist[i][j] < dist[lowestSource][lowestDestination] {
								lowestSource = i
								lowestDestination = j
							}
						}
					}
				}
			}

			if prev[lowestSource][lowestDestination] == -1 {
				return nil, errors.New("no longest path found")
			}

			path := []string{}
			opinions := []subjectivelogic.Opinion{}
			targetSource := reverseNodes[lowestSource]
			targetDestination := reverseNodes[lowestDestination]
			targetKey := targetSource + ":" + targetDestination

			for lowestSource != lowestDestination {
				target := reverseNodes[lowestDestination]
				lowestDestination = prev[lowestSource][lowestDestination]
				key := reverseNodes[lowestDestination] + ":" + target
				if len(current[key].opinions) != 1 {
					return nil, errors.New("too many opinions for discounting")
				}
				path = append([]string{key}, path...)
				opinions = append([]subjectivelogic.Opinion{current[key].opinions[0]}, opinions...)
				delete(current, key)
			}

			var discounted subjectivelogic.Opinion

			switch len(opinions) {
			case 0:
				return nil, errors.New("an out of opinions without result")
				break L

			case 1:
				results[scope] = &opinions[0]
				break L

			case 2:
				tmp, err := subjectivelogic.TrustDiscounting(&opinions[0], &opinions[1])
				if err != nil {
					return nil, errors.New("subjective logic error")
				}
				discounted = tmp

			default:
				tmp, err := subjectivelogic.MultiEdgeTrustDisc(opinions)
				if err != nil {
					return nil, errors.New("subjective logic error")
				}
				discounted = tmp
			}

			entry, exists := current[targetKey]
			if !exists {
				current[targetKey] = CurrentEntry{
					source:      targetSource,
					destination: targetDestination,
					opinions:    []subjectivelogic.Opinion{discounted},
				}
			} else {
				entry.opinions = append(entry.opinions, discounted)
				current[targetKey] = entry
			}
		}
	}

	return results, nil
}

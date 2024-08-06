package trustassessment

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
)

/*
The ResultEntry type is an internal, type representing output elements for TAS_NOTIFY or TAS_TA_RESPONSE messages.
It is agnostic from the auto-generated structs used for compiling actual results.
*/
type ResultEntry struct {
	TmiID        string
	Propositions []Proposition
}

type Proposition struct {
	PropositionID string
	ATL           subjectivelogic.QueryableOpinion
	PP            float64
	TrustDecision core.TrustDecision
}

/*
The NewPropositionEntry function is a helper function that creates proposition structs based on an existing AtlResultSets and a proposition ID.
*/
func NewPropositionEntry(set core.AtlResultSet, propositionID string) Proposition {
	return Proposition{
		PropositionID: propositionID,
		ATL:           set.ATLs()[propositionID],
		PP:            set.ProjectedProbabilities()[propositionID],
		TrustDecision: set.TrustDecisions()[propositionID],
	}
}

/*
toMsgStruct takes an internal representation of a TMI/proposition result and converts into message struct auto-generated from JSON Schema.
Result variant.
*/
func (r ResultEntry) toResultMsgStruct() tasmsg.Result {

	propositions := make([]tasmsg.ResultProposition, 0)

	for _, proposition := range r.Propositions {

		var tdValue *bool = nil
		if proposition.TrustDecision == core.TRUSTWORTHY {
			value := true
			tdValue = &value
		} else if proposition.TrustDecision == core.NOT_TRUSTWORTHY {
			value := false
			tdValue = &value
		}

		atl := make([]tasmsg.FluffyActualTrustworthinessLevel, 0)
		baseRate := proposition.ATL.BaseRate()
		belief := proposition.ATL.Belief()
		disbelief := proposition.ATL.Disbelief()
		uncertainty := proposition.ATL.Uncertainty()

		atl = append(atl, tasmsg.FluffyActualTrustworthinessLevel{
			Output: tasmsg.FluffyOutput{
				BaseRate:    &baseRate,
				Belief:      &belief,
				Disbelief:   &disbelief,
				Uncertainty: &uncertainty,
			},
			Type: tasmsg.SubjectiveLogicOpinion,
		})

		projectedProbability := proposition.PP
		atl = append(atl, tasmsg.FluffyActualTrustworthinessLevel{
			Output: tasmsg.FluffyOutput{
				Value: &projectedProbability,
			},
			Type: tasmsg.ProjectedProbability,
		})

		propositions = append(propositions, tasmsg.ResultProposition{
			ActualTrustworthinessLevel: atl,
			PropositionID:              proposition.PropositionID,
			TrustDecision:              tdValue,
		})
	}

	return tasmsg.Result{
		ID:           r.TmiID,
		Propositions: propositions,
	}

}

/*
toMsgStruct takes an internal representation of a TMI/proposition result and converts into message struct auto-generated from JSON Schema.
Update variant.
*/
func (r ResultEntry) toUpdateMsgStruct() tasmsg.Update {

	propositions := make([]tasmsg.UpdateProposition, 0)

	for _, proposition := range r.Propositions {

		var tdValue *bool = nil
		if proposition.TrustDecision == core.TRUSTWORTHY {
			value := true
			tdValue = &value
		} else if proposition.TrustDecision == core.NOT_TRUSTWORTHY {
			value := false
			tdValue = &value
		}

		atl := make([]tasmsg.PurpleActualTrustworthinessLevel, 0)
		baseRate := proposition.ATL.BaseRate()
		belief := proposition.ATL.Belief()
		disbelief := proposition.ATL.Disbelief()
		uncertainty := proposition.ATL.Uncertainty()

		atl = append(atl, tasmsg.PurpleActualTrustworthinessLevel{
			Output: tasmsg.PurpleOutput{
				BaseRate:    &baseRate,
				Belief:      &belief,
				Disbelief:   &disbelief,
				Uncertainty: &uncertainty,
			},
			Type: tasmsg.SubjectiveLogicOpinion,
		})

		projectedProbability := proposition.PP
		atl = append(atl, tasmsg.PurpleActualTrustworthinessLevel{
			Output: tasmsg.PurpleOutput{
				Value: &projectedProbability,
			},
			Type: tasmsg.ProjectedProbability,
		})

		propositions = append(propositions, tasmsg.UpdateProposition{
			ActualTrustworthinessLevel: atl,
			PropositionID:              proposition.PropositionID,
			TrustDecision:              tdValue,
		})
	}

	return tasmsg.Update{
		ID:           r.TmiID,
		Propositions: propositions,
	}

}

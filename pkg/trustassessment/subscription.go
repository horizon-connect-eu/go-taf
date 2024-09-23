package trustassessment

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	"math"
)

type Trigger string

const (
	ACTUAL_TRUSTWORTHINESS_LEVEL Trigger = "ACTUAL_TRUSTWORTHINESS_LEVEL"
	TRUST_DECISION               Trigger = "TRUST_DECISION"
)

type Subscription interface {
	Trigger() Trigger
	SubscriptionID() string
	SessionID() string
	HandleUpdate(old core.AtlResultSet, new core.AtlResultSet) []ResultEntry
	SubscriberTopic() string
}

type SubscriptionInstance struct {
	subscriptionID  string
	subscriberTopic string
	sessionID       string
	//tmiID->bool
	filter  map[string]bool
	trigger Trigger
}

func NewSubscription(subscriptionID string, sessionID string, subscriberTopic string, filterList []string, trigger Trigger) Subscription {

	filter := make(map[string]bool)
	for _, item := range filterList {
		filter[item] = true
	}

	return &SubscriptionInstance{
		subscriptionID:  subscriptionID,
		subscriberTopic: subscriberTopic,
		sessionID:       sessionID,
		filter:          filter,
		trigger:         trigger,
	}
}

func (s *SubscriptionInstance) Trigger() Trigger {
	return s.trigger
}
func (s *SubscriptionInstance) SubscriptionID() string {
	return s.subscriptionID
}
func (s *SubscriptionInstance) SessionID() string {
	return s.sessionID
}

func (s *SubscriptionInstance) SubscriberTopic() string {
	return s.subscriberTopic
}

func (s *SubscriptionInstance) HandleUpdate(oldATLs core.AtlResultSet, newATLs core.AtlResultSet) []ResultEntry {
	result := make([]ResultEntry, 0)
	propositions := make([]Proposition, 0)

	if oldATLs.SessionID() != newATLs.SessionID() || oldATLs.TmiID() != newATLs.TmiID() {
		return result
	}
	if len(s.filter) > 0 {
		_, exists := s.filter[newATLs.TmiID()]
		if !exists {
			return result
		}
	}

	switch s.trigger {
	case ACTUAL_TRUSTWORTHINESS_LEVEL:
		for propositionID, newOpinion := range newATLs.ATLs() {
			oldOpinion, exists := oldATLs.ATLs()[propositionID]
			if !exists {
				//Proposition has not yet existed, so add as changed!
				propositions = append(propositions, NewPropositionEntry(newATLs, propositionID))
			} else {
				if !areIdenticalSubjectiveLogicOpinions(oldOpinion, newOpinion) {
					//There is a change in the ATL, so add as changed.
					propositions = append(propositions, NewPropositionEntry(newATLs, propositionID))
				}
			}
		}
	case TRUST_DECISION:
		for propositionID, newTD := range newATLs.TrustDecisions() {
			oldTD, exists := oldATLs.TrustDecisions()[propositionID]
			if !exists {
				//Proposition has not yet existed, so add as changed!
				propositions = append(propositions, NewPropositionEntry(newATLs, propositionID))
			}
			if oldTD != newTD {
				//There is a change in the Trust Decision, so add as changed.
				propositions = append(propositions, NewPropositionEntry(newATLs, propositionID))
			}
		}
	default:
		//Nothing to do
	}

	if len(propositions) > 0 {
		result = append(result, ResultEntry{
			TmiID:        newATLs.TmiID(),
			Propositions: propositions,
		})
	}
	return result
}

/*
precision defines the maximum deviance each value of an Opinion can have for the Opinion to still be regarded as a valid Binomial Opinion.
*/
const precision float64 = 0.000000000001

// TODO copied from SL library, should later be replaced by library directly
func areIdenticalSubjectiveLogicOpinions(opinion1 subjectivelogic.QueryableOpinion, opinion2 subjectivelogic.QueryableOpinion) bool {
	return math.Abs(opinion1.Belief()-opinion2.Belief()) < precision &&
		math.Abs(opinion1.Disbelief()-opinion2.Disbelief()) < precision &&
		math.Abs(opinion1.Uncertainty()-opinion2.Uncertainty()) < precision &&
		math.Abs(opinion1.BaseRate()-opinion2.BaseRate()) < precision
}

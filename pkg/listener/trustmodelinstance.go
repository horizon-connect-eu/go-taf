package listener

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"time"
)

type TrustModelInstanceListener interface {
	OnTrustModelInstanceSpawned(event TrustModelInstanceSpawnedEvent)
	OnTrustModelInstanceUpdated(event TrustModelInstanceUpdatedEvent)
	OnTrustModelInstanceDeleted(event TrustModelInstanceDeletedEvent)
}

type TrustModelInstanceSpawnedEvent struct {
	EventType   EventType
	Timestamp   time.Time
	ID          string
	FullTMI     string
	Template    core.TrustModelTemplate
	Version     int
	Fingerprint uint32
	Structure   trustmodelstructure.TrustGraphStructure
	Values      map[string][]trustmodelstructure.TrustRelationship
	RTLs        map[string]subjectivelogic.QueryableOpinion
}

func NewTrustModelInstanceSpawnedEvent(ID string, fullTMI string, template core.TrustModelTemplate, version int, fingerprint uint32, structure trustmodelstructure.TrustGraphStructure, values map[string][]trustmodelstructure.TrustRelationship, RTLs map[string]subjectivelogic.QueryableOpinion) TrustModelInstanceSpawnedEvent {
	return TrustModelInstanceSpawnedEvent{
		Timestamp: time.Now(),
		EventType: TRUST_MODEL_INSTANCE_SPAWNED,
		ID:        ID, FullTMI: fullTMI, Template: template, Version: version, Fingerprint: fingerprint, Structure: structure, Values: values, RTLs: RTLs}
}
func (e TrustModelInstanceSpawnedEvent) Event() EventType {
	return e.EventType
}

type TrustModelInstanceUpdatedEvent struct {
	EventType   EventType
	Timestamp   time.Time
	ID          string
	FullTMI     string
	Template    core.TrustModelTemplate
	Version     int
	Fingerprint uint32
	Structure   trustmodelstructure.TrustGraphStructure
	Values      map[string][]trustmodelstructure.TrustRelationship
	RTLs        map[string]subjectivelogic.QueryableOpinion
}

func NewTrustModelInstanceUpdatedEvent(ID string, fullTMI string, template core.TrustModelTemplate, version int, fingerprint uint32, structure trustmodelstructure.TrustGraphStructure, values map[string][]trustmodelstructure.TrustRelationship, RTLs map[string]subjectivelogic.QueryableOpinion) TrustModelInstanceUpdatedEvent {
	return TrustModelInstanceUpdatedEvent{
		Timestamp: time.Now(),
		EventType: TRUST_MODEL_INSTANCE_UPDATED,
		ID:        ID, FullTMI: fullTMI, Template: template, Version: version, Fingerprint: fingerprint, Structure: structure, Values: values, RTLs: RTLs}
}
func (e TrustModelInstanceUpdatedEvent) Event() EventType {
	return e.EventType
}

type TrustModelInstanceDeletedEvent struct {
	EventType EventType
	Timestamp time.Time
	FullTMI   string
}

func NewTrustModelInstanceDeletedEvent(fullTMI string) TrustModelInstanceDeletedEvent {
	return TrustModelInstanceDeletedEvent{
		Timestamp: time.Now(),
		EventType: TRUST_MODEL_INSTANCE_DELETED,
		FullTMI:   fullTMI}
}
func (e TrustModelInstanceDeletedEvent) Event() EventType {
	return e.EventType
}

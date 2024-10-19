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

type TrustModelInstanceUpdatedEvent struct {
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

type TrustModelInstanceDeletedEvent struct {
	Timestamp time.Time
	FullTMI   string
}

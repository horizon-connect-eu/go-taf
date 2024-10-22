package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/listener"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"log/slog"
	"net/http"
	"strconv"
)

/*
The State struct is a copy of the TAF internal state, built from an event stream sent from the TAF.
This copy can then be used by the WEB UI to give read access to the recreated TAF state.
This includes:
  - stream of internal events sent from the TAF
  - state of TMIs
*/
type State struct {
	eventLog           []listener.ListenerEvent
	logger             *slog.Logger
	tmis               map[string]*tmiMetaState
	tmisLatestVersions map[string]int
}

func NewState(logger *slog.Logger) *State {
	return &State{
		logger:             logger,
		eventLog:           make([]listener.ListenerEvent, 0),
		tmis:               make(map[string]*tmiMetaState),
		tmisLatestVersions: make(map[string]int),
	}
}

type tmiState struct {
	Version     int
	Fingerprint uint32
	Structure   trustmodelstructure.TrustGraphStructure
	Values      map[string][]trustmodelstructure.TrustRelationship
	RTLs        map[string]subjectivelogic.QueryableOpinion
}

type tmiMetaState struct {
	ID            string
	FullTMI       string
	IsActive      bool
	LatestVersion int
	Update        map[int]core.Update
	States        map[int]tmiState
	Template      core.TrustModelTemplate
	ATLs          map[int]core.AtlResultSet
}

func (s *State) Handle(incomingEvents chan listener.ListenerEvent) {
	for {
		select {
		case evt := <-incomingEvents:
			s.eventLog = append([]listener.ListenerEvent{evt}, s.eventLog...)
			switch event := evt.(type) {
			case listener.ATLRemovedEvent:
				s.logger.Info("ATLRemovedEvent")
			case listener.ATLUpdatedEvent:
				s.logger.Info("ATLUpdatedEvent")
				s.handleATLUpdatedEvent(event)
			case listener.TrustModelInstanceSpawnedEvent:
				s.logger.Info("TrustModelInstanceSpawnedEvent")
				s.handleTMISpawned(event)
			case listener.TrustModelInstanceUpdatedEvent:
				s.logger.Info("TrustModelInstanceUpdatedEvent")
				s.handleTMIUpdated(event)
			case listener.TrustModelInstanceDeletedEvent:
				s.logger.Info("TrustModelInstanceDeletedEvent")
				s.handleTMIDeleted(event)
			case listener.SessionCreatedEvent:
				s.logger.Info("SessionCreatedEvent")
			case listener.SessionTorndownEvent:
				s.logger.Info("SessionTorndownEvent")
			default:
				util.UNUSED(event)
			}
		}
	}
}

func (s *State) handleATLUpdatedEvent(event listener.ATLUpdatedEvent) {
	fullTMI := event.FullTMI
	_, exists := s.tmis[fullTMI]
	if !exists {
		return
	} else {
		s.logger.Warn(fmt.Sprintf("%+v", event.NewATLs))
	}
	s.tmis[fullTMI].ATLs[event.NewATLs.Version()] = event.NewATLs
}

func (s *State) handleTMISpawned(event listener.TrustModelInstanceSpawnedEvent) {
	fullTMI := event.FullTMI

	s.tmis[fullTMI] = &tmiMetaState{
		IsActive:      true,
		LatestVersion: 0,
		Update:        make(map[int]core.Update),
		States:        make(map[int]tmiState),
		Template:      event.Template,
		ID:            event.ID,
		FullTMI:       event.FullTMI,
		ATLs:          make(map[int]core.AtlResultSet),
	}
	s.tmis[fullTMI].States[event.Version] = tmiState{
		Version:     event.Version,
		Fingerprint: event.Fingerprint,
		Structure:   event.Structure,
		Values:      event.Values,
		RTLs:        event.RTLs,
	}

}

func (s *State) handleTMIUpdated(event listener.TrustModelInstanceUpdatedEvent) {
	fullTMI := event.FullTMI
	s.tmis[fullTMI].States[event.Version] = tmiState{
		Version:     event.Version,
		Fingerprint: event.Fingerprint,
		Structure:   event.Structure,
		Values:      event.Values,
		RTLs:        event.RTLs,
	}
	s.tmis[fullTMI].Update[event.Version] = event.Update
	s.tmis[fullTMI].LatestVersion = event.Version
}

func (s *State) handleTMIDeleted(event listener.TrustModelInstanceDeletedEvent) {

}

func (s *State) getEventLog(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, s.eventLog)
}

func (s *State) getTMI(ctx *gin.Context) {
	fullTMI := core.MergeFullTMIIdentifier(ctx.Param("client"), ctx.Param("session"), ctx.Param("tmt"), ctx.Param("tmiID"))
	_, exists := s.tmis[fullTMI]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
	} else {
		ctx.Redirect(http.StatusFound, ctx.Request.URL.Path+"/latest")
	}
}

func (s *State) getVersionTMI(ctx *gin.Context) {
	fullTMI := core.MergeFullTMIIdentifier(ctx.Param("client"), ctx.Param("session"), ctx.Param("tmt"), ctx.Param("tmiID"))
	tmi, exists := s.tmis[fullTMI]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
	} else {
		version, err := strconv.Atoi(ctx.Param("version"))
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"code": "VERSION_NOT_FOUND"})
			return
		}
		if _, exists := s.tmis[fullTMI].States[version]; !exists {
			ctx.JSON(http.StatusNotFound, gin.H{"code": "VERSION_NOT_FOUND"})
			return
		} else {
			res := gin.H{
				"id":       tmi.ID,
				"fullTMI":  tmi.FullTMI,
				"active":   tmi.IsActive,
				"template": tmi.Template.Identifier(),
				"state":    tmi.States[version],
				"updates":  tmi.Update[version],
			}

			atls, atlExist := tmi.ATLs[version]
			if !atlExist {
				res["atls"] = nil
			} else {
				res["atls"] = atls
			}
			ctx.JSON(http.StatusOK, res)
		}
	}
}

func (s *State) getTMILatest(ctx *gin.Context) {
	fullTMI := core.MergeFullTMIIdentifier(ctx.Param("client"), ctx.Param("session"), ctx.Param("tmt"), ctx.Param("tmiID"))
	tmi, exists := s.tmis[fullTMI]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
	} else {
		latestVersion := tmi.LatestVersion
		res := gin.H{
			"id":       tmi.ID,
			"fullTMI":  tmi.FullTMI,
			"active":   tmi.IsActive,
			"template": tmi.Template.Identifier(),
			"state":    tmi.States[latestVersion],
			"updates":  tmi.Update[latestVersion],
		}

		atls, atlExist := tmi.ATLs[latestVersion]
		if !atlExist {
			res["atls"] = nil
		} else {
			res["atls"] = atls
		}
		ctx.JSON(http.StatusOK, res)

	}
}

func (s *State) getAllTMIs(ctx *gin.Context) {

	tmis := make(map[string]interface{})
	i := 0
	for fullTMI := range s.tmis {
		tmis[fullTMI] = struct {
			Id            string `json:"id"`
			FullTMI       string `json:"fullTMI"`
			Active        bool   `json:"active"`
			Template      string `json:"template"`
			LatestVersion int    `json:"latestVersion"`
		}{
			s.tmis[fullTMI].ID,
			fullTMI,
			s.tmis[fullTMI].IsActive,
			s.tmis[fullTMI].Template.Identifier(),
			s.tmis[fullTMI].LatestVersion,
		}
		i++
	}
	ctx.JSON(http.StatusOK, tmis)
}

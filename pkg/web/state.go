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
	eventLog []listener.ListenerEvent
	logger   *slog.Logger
	tmis     map[string]*tmiMetaState
	sessions map[string]*sessionState
}

func NewState(logger *slog.Logger) *State {
	return &State{
		logger:   logger,
		eventLog: make([]listener.ListenerEvent, 0),
		tmis:     make(map[string]*tmiMetaState),
		sessions: make(map[string]*sessionState),
	}
}

/*
PAGE_SIZE defines the maximum amount of items shown with pagination.
*/
const PAGE_SIZE = 100

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

type sessionState struct {
	Client   string
	IsActive bool
	TMIs     []string
	Template string
}

func (s *State) Handle(incomingEvents chan listener.ListenerEvent) {
	for {
		select {
		case evt := <-incomingEvents:
			s.eventLog = append(s.eventLog, evt)
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
				s.handleSessionCreatedEvent(event)
			case listener.SessionTorndownEvent:
				s.logger.Info("SessionTorndownEvent")
				s.handleSessionTorndownEvent(event)
			default:
				util.UNUSED(event)
			}
		}
	}
}

func (s *State) handleSessionCreatedEvent(event listener.SessionCreatedEvent) {
	s.sessions[event.SessionID] = &sessionState{
		Client:   event.ClientID,
		IsActive: true,
		TMIs:     make([]string, 0),
		Template: event.TrustModelTemplate,
	}
}

func (s *State) handleSessionTorndownEvent(event listener.SessionTorndownEvent) {
	if _, exists := s.sessions[event.SessionID]; exists {
		s.sessions[event.SessionID].IsActive = false
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

	_, sessionID, _, _ := core.SplitFullTMIIdentifier(fullTMI)
	if _, exists := s.sessions[sessionID]; exists {
		s.sessions[sessionID].TMIs = append(s.sessions[sessionID].TMIs, fullTMI)
	}

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

func (s *State) getSessions(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, s.sessions)
}

func (s *State) getFullEventLog(ctx *gin.Context) {
	fullLog := make([]map[int]interface{}, len(s.eventLog))

	logLength := len(s.eventLog)
	for i, entry := range s.eventLog {
		fullLog[logLength-i-1] = map[int]interface{}{
			i: entry,
		}
	}
	ctx.JSON(http.StatusOK, fullLog)
}

func (s *State) getEventLogPage(ctx *gin.Context) {

	var cursor int
	rawCursor, exists := ctx.GetQuery("cursor")
	if !exists {
		ctx.Redirect(http.StatusFound, ctx.Request.URL.Path+"?cursor=0")
		return
	} else {
		c, err := strconv.Atoi(rawCursor)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"code": "INVALID_CURSOR"})
			return
		} else {
			cursor = c
		}
	}

	latestIdx := max(0, len(s.eventLog)-1)

	lower := max(0, cursor)
	upper := min(latestIdx, cursor+PAGE_SIZE)

	if upper-lower == 0 {
		ctx.JSON(http.StatusOK, make([]interface{}, 0))
		return
	} else if lower > upper {
		ctx.JSON(http.StatusOK, gin.H{"code": "INVALID_CURSOR"})
		return
	}
	log := make([]map[int]interface{}, upper-lower)

	for i := lower; i <= upper-1; i++ {
		log[upper-i-1] = map[int]interface{}{
			i: s.eventLog[i],
		}
	}

	res := gin.H{"page": log}
	prev := max(lower-PAGE_SIZE, 0)
	if cursor != prev {
		res["previous"] = prev
	}
	next := min(lower+PAGE_SIZE, latestIdx)
	if cursor != next {
		res["next"] = next
	}
	ctx.JSON(http.StatusOK, res)
}

func (s *State) getLatestEventLogPage(ctx *gin.Context) {

	latestIdx := max(0, len(s.eventLog)-1)
	lower := max(0, (latestIdx - PAGE_SIZE))

	log := make([]map[int]interface{}, latestIdx-lower)

	if latestIdx-lower == 0 {
		ctx.JSON(http.StatusOK, log)
		return
	}

	for i := lower; i <= latestIdx-1; i++ {
		log[latestIdx-i-1] = map[int]interface{}{
			i: s.eventLog[i],
		}
	}
	ctx.JSON(http.StatusOK, log)
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

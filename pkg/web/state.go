package web

import (
	"github.com/gin-gonic/gin"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/listener"
	"log/slog"
	"net/http"
)

type State struct {
	eventLog []listener.ListenerEvent
	logger   *slog.Logger
	tmis     map[string]interface{}
}

func NewState(logger *slog.Logger) *State {
	return &State{
		logger:   logger,
		eventLog: make([]listener.ListenerEvent, 0),
		tmis:     make(map[string]interface{}),
	}
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
			case listener.TrustModelInstanceSpawnedEvent:
				s.logger.Info("TrustModelInstanceSpawnedEvent")
				s.tmis[event.FullTMI] = event
			case listener.TrustModelInstanceUpdatedEvent:
				s.logger.Info("TrustModelInstanceUpdatedEvent")
				s.tmis[event.FullTMI] = event
			case listener.TrustModelInstanceDeletedEvent:
				s.logger.Info("TrustModelInstanceDeletedEvent")
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

func (s *State) getEventLog(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, s.eventLog)
}

func (s *State) getTMI(ctx *gin.Context) {
	fullTMI := core.MergeFullTMIIdentifier(ctx.Param("client"), ctx.Param("session"), ctx.Param("tmt"), ctx.Param("tmiID"))
	tmi, exists := s.tmis[fullTMI]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
	} else {
		ctx.IndentedJSON(http.StatusOK, tmi)
	}
}

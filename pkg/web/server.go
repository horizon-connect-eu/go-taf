package web

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/listener"
	"log/slog"
	"net/http"
)

//https://www.jetbrains.com/guide/go/tutorials/rest_api_series/gin/

type Webserver struct {
	tafContext      core.TafContext
	channels        core.TafChannels
	logger          *slog.Logger
	router          *gin.Engine
	listenerChannel chan interface{}
}

var eventLog = make([]interface{}, 0)
var tmis = make(map[string]interface{})

func New(tafContext core.TafContext, channels core.TafChannels) (*Webserver, error) {
	return &Webserver{
		tafContext:      tafContext,
		channels:        channels,
		logger:          logging.CreateChildLogger(tafContext.Logger, "WEB-UI"),
		listenerChannel: make(chan interface{}),
	}, nil
}

func (s *Webserver) Run() {
	gin.SetMode(gin.ReleaseMode) //Disable Gin-specific logging output
	s.router = gin.New()         //Create a non-default router without request logging
	s.router.Use(gin.Recovery())
	s.router.GET("/trustsources", getTrustSources)
	s.router.GET("/events", getEventLog)
	s.router.GET("/tmis/:client/:session/:tmt/:tmiID", s.getTMI)
	go s.router.Run(fmt.Sprintf(":%d", s.tafContext.Configuration.WebUI.Port))

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(s.tafContext.Context); err != nil {
			return
		}
		select {
		case <-s.tafContext.Context.Done():
			return
		case evt := <-s.listenerChannel:
			eventLog = append(eventLog, evt)
			switch event := evt.(type) {
			case listener.ATLRemovedEvent:
				s.logger.Info("ATLRemovedEvent")
			case listener.ATLUpdatedEvent:
				s.logger.Info("ATLUpdatedEvent")
			case listener.TrustModelInstanceSpawnedEvent:
				s.logger.Info("TrustModelInstanceSpawnedEvent")
				tmis[event.FullTMI] = event
			case listener.TrustModelInstanceUpdatedEvent:
				s.logger.Info("TrustModelInstanceUpdatedEvent")
				tmis[event.FullTMI] = event
			case listener.TrustModelInstanceDeletedEvent:
				s.logger.Info("TrustModelInstanceDeletedEvent")
			case listener.SessionCreatedEvent:
				s.logger.Info("SessionCreatedEvent")
			case listener.SessionDeletedEvent:
				s.logger.Info("SessionDeletedEvent")
			default:
				util.UNUSED(event)
			}
		}
	}
}

func getTrustSources(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, map[string]interface{}{})
}
func getEventLog(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, eventLog)
}

func (s *Webserver) getTMI(ctx *gin.Context) {
	fullTMI := core.MergeFullTMIIdentifier(ctx.Param("client"), ctx.Param("session"), ctx.Param("tmt"), ctx.Param("tmiID"))
	tmi, exists := tmis[fullTMI]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
	} else {
		ctx.IndentedJSON(http.StatusOK, tmi)
	}
}

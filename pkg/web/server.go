package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/listener"
	"github.com/vs-uulm/go-taf/pkg/manager"
	"log/slog"
	"net/http"
)

//https://www.jetbrains.com/guide/go/tutorials/rest_api_series/gin/

type Webserver struct {
	tafContext      core.TafContext
	managers        manager.TafManagers
	logger          *slog.Logger
	router          *gin.Engine
	listenerChannel chan listener.ListenerEvent
	State           *State
}

func New(tafContext core.TafContext) (*Webserver, error) {
	logger := logging.CreateChildLogger(tafContext.Logger, "WEB-UI")
	return &Webserver{
		tafContext:      tafContext,
		logger:          logger,
		listenerChannel: make(chan listener.ListenerEvent),
		State:           NewState(logger),
	}, nil
}

func (s *Webserver) SetManagers(managers manager.TafManagers) {
	s.managers = managers
}

func (s *Webserver) Run() {
	go s.State.Handle(s.listenerChannel)

	gin.SetMode(gin.ReleaseMode) //Disable Gin-specific logging output
	s.router = gin.New()         //Create a non-default router without request logging
	s.router.Use(gin.Recovery())
	s.router.GET("/trustsources", s.getTrustSources)
	s.router.GET("/events", s.State.getEventLog)
	s.router.GET("/tmis/:client/:session/:tmt/:tmiID", s.State.getTMI)
	s.router.Run(fmt.Sprintf(":%d", s.tafContext.Configuration.WebUI.Port))
}

func (s *Webserver) getTrustSources(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, s.managers.TMM.GetAllTMTs())
}

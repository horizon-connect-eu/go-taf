package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/version"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/listener"
	"github.com/vs-uulm/go-taf/pkg/manager"
	"log/slog"
	"net/http"
)

//https://www.jetbrains.com/guide/go/tutorials/rest_api_series/gin/

type Webserver struct {
	tafContext      core.TafContext
	logger          *slog.Logger
	router          *gin.Engine
	listenerChannel chan listener.ListenerEvent
	state           *State
	tmts            map[string]interface{}
	trustSources    map[string]map[string]bool
}

func New(tafContext core.TafContext) (*Webserver, error) {
	logger := logging.CreateChildLogger(tafContext.Logger, "WEB-UI")
	return &Webserver{
		tafContext:      tafContext,
		logger:          logger,
		listenerChannel: make(chan listener.ListenerEvent),
		state:           NewState(logger),
		tmts:            make(map[string]interface{}),
		trustSources:    make(map[string]map[string]bool),
	}, nil
}

func (s *Webserver) Run() {
	go s.state.Handle(s.listenerChannel)

	gin.SetMode(gin.ReleaseMode) //Disable Gin-specific logging output
	s.router = gin.New()         //Create a non-default router without request logging
	s.router.Use(gin.Recovery())
	//	s.router.LoadHTMLGlob("res/templates/*")
	s.router.GET("/events", s.state.getEventLogPage)
	s.router.GET("/events/latest", s.state.getLatestEventLogPage)
	s.router.GET("/events/all", s.state.getFullEventLog)
	s.router.GET("/info", s.getInfo)
	s.router.GET("/sessions", s.state.getSessions)
	s.router.GET("/tmis/", s.state.getAllTMIs)
	s.router.GET("/tmis/:client/:session/:tmt/:tmiID", s.state.getTMI)
	s.router.GET("/tmis/:client/:session/:tmt/:tmiID/latest", s.state.getTMILatest)
	s.router.GET("/tmis/:client/:session/:tmt/:tmiID/updates", s.state.getTMIUpdates)
	s.router.GET("/tmis/:client/:session/:tmt/:tmiID/all", s.state.getTMIFull)
	s.router.GET("/tmis/:client/:session/:tmt/:tmiID/:version", s.state.getVersionTMI)
	s.router.GET("/trustmodels/:tmt-identifier", s.getTrustModel)
	s.router.GET("/trustsources", s.getTrustSources)
	s.router.GET("/trustmodels", s.getTrustModels)
	s.router.Run(fmt.Sprintf(":%d", s.tafContext.Configuration.WebUI.Port))
}

func (s *Webserver) getTrustModels(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, s.tmts)
}

func (s *Webserver) getTrustSources(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, s.trustSources)
}

func (s *Webserver) getTrustModel(ctx *gin.Context) {
	tmt, exists := s.tmts[ctx.Param("tmt-identifier")]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
	} else {
		ctx.IndentedJSON(http.StatusOK, tmt)
	}
}

func (s *Webserver) getInfo(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"Version": version.Version, "Build": version.Build, "Configuration": s.tafContext.Configuration})
}

func (s *Webserver) SetManagers(managers manager.TafManagers) {

	for _, tmt := range managers.TMM.GetAllTMTs() {

		evidence := make(map[string]map[string]bool)
		for _, evidenceType := range tmt.EvidenceTypes() {
			if evidence[evidenceType.Source().String()] == nil {
				evidence[evidenceType.Source().String()] = make(map[string]bool)
			}
			evidence[evidenceType.Source().String()][evidenceType.String()] = true

			if s.trustSources[evidenceType.Source().String()] == nil {
				s.trustSources[evidenceType.Source().String()] = make(map[string]bool)
			}
			s.trustSources[evidenceType.Source().String()][evidenceType.String()] = true
		}

		s.tmts[tmt.Identifier()] = struct {
			Description   string
			Name          string
			Version       string
			EvidenceTypes map[string]map[string]bool
		}{
			Description:   tmt.Description(),
			Name:          tmt.TemplateName(),
			Version:       tmt.Version(),
			EvidenceTypes: evidence,
		}
	}
}

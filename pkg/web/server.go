package web

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	logging "github.com/horizon-connect-eu/go-taf/internal/logger"
	"github.com/horizon-connect-eu/go-taf/internal/version"
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	"github.com/horizon-connect-eu/go-taf/pkg/listener"
	"github.com/horizon-connect-eu/go-taf/pkg/manager"
)

//go:embed frontend/dist
var webFrontend embed.FS

//https://www.jetbrains.com/guide/go/tutorials/rest_api_series/gin/

type Webserver struct {
	tafContext       core.TafContext
	logger           *slog.Logger
	router           *gin.Engine
	listenerChannel  chan listener.ListenerEvent
	websocketChannel chan WebSocketEvent
	state            *State
	tmts             map[string]interface{}
	trustSources     map[string]map[string]bool
}

func New(tafContext core.TafContext) (*Webserver, error) {
	logger := logging.CreateChildLogger(tafContext.Logger, "WEB-UI")
	return &Webserver{
		tafContext:       tafContext,
		logger:           logger,
		listenerChannel:  make(chan listener.ListenerEvent),
		websocketChannel: make(chan WebSocketEvent),
		state:            NewState(logger),
		tmts:             make(map[string]interface{}),
		trustSources:     make(map[string]map[string]bool),
	}, nil
}

type WebSocketConnectedEvent struct {
	Timestamp  time.Time
	Connection *websocket.Conn
}

func (s *Webserver) Run() {
	go s.state.Handle(s.listenerChannel, s.websocketChannel)

	staticFS := fs.FS(webFrontend)
	frontendDir, err := fs.Sub(staticFS, "frontend/dist")
	if err != nil {
		panic(err)
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	gin.SetMode(gin.ReleaseMode) //Disable Gin-specific logging output
	s.router = gin.New()         //Create a non-default router without request logging
	s.router.Use(func(c *gin.Context) {
		s.logger.Info("gin request", "uri", c.Request.RequestURI)
		c.Next()
	})

	s.router.Use(gin.Recovery())
	s.router.StaticFS("/ui", http.FS(frontendDir))
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/ui")
	})
	s.router.GET("/ws", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		defer ws.Close()
		s.websocketChannel <- WebSocketEvent{Socket: ws, Type: REGISTER}

		for {
			messageType, msg, err := ws.ReadMessage()
			if err != nil {
				s.websocketChannel <- WebSocketEvent{Socket: ws, Type: UNREGISTER}
				break
			}

			if messageType == websocket.BinaryMessage || messageType == websocket.TextMessage {
				s.logger.Info("received ws message", "msg", msg)
			}
		}
	})

	//	s.router.LoadHTMLGlob("res/templates/*")
	s.router.GET("/api/events", s.state.getEventLogPage)
	s.router.GET("/api/events/latest", s.state.getLatestEventLogPage)
	s.router.GET("/api/events/all", s.state.getFullEventLog)
	s.router.GET("/api/info", s.getInfo)
	s.router.GET("/api/sessions", s.state.getSessions)
	s.router.GET("/api/tmis/", s.state.getAllTMIs)
	s.router.GET("/api/tmis/:client/:session/:tmt/:tmiID", s.state.getTMI)
	s.router.GET("/api/tmis/:client/:session/:tmt/:tmiID/latest", s.state.getTMILatest)
	s.router.GET("/api/tmis/:client/:session/:tmt/:tmiID/updates", s.state.getTMIUpdates)
	s.router.GET("/api/tmis/:client/:session/:tmt/:tmiID/all", s.state.getTMIFull)
	s.router.GET("/api/tmis/:client/:session/:tmt/:tmiID/:version", s.state.getVersionTMI)
	s.router.GET("/api/trustmodels/:tmt-identifier", s.getTrustModel)
	s.router.GET("/api/trustsources", s.getTrustSources)
	s.router.GET("/api/trustmodels", s.getTrustModels)
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

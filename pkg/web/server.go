package web

import (
	"github.com/gin-gonic/gin"
	"github.com/vs-uulm/go-taf/pkg/core"
	"net/http"
)

//https://www.jetbrains.com/guide/go/tutorials/rest_api_series/gin/

type Webserver struct {
	tafContext core.TafContext
	channels   core.TafChannels
	router     *gin.Engine
}

func New(tafContext core.TafContext, channels core.TafChannels) (*Webserver, error) {
	return &Webserver{
		tafContext: tafContext,
		channels:   channels,
	}, nil
}

func (s *Webserver) Run() {
	gin.SetMode(gin.ReleaseMode) //Disable Gin-specific logging output
	s.router = gin.New()         //Create a non-default router without request logging
	s.router.Use(gin.Recovery())
	s.router.GET("/trustsources", getTrustSources)
	s.router.Run("localhost:7778")
}

func getTrustSources(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, map[string]interface{}{})
}

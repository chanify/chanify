package core

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Core struct {
}

func New() *Core {
	gin.SetMode(gin.ReleaseMode)
	return &Core{}
}

func (c *Core) Close() {
}

func (c *Core) APIHandler() http.Handler {
	r := gin.New()
	r.Use(loggerMiddleware)
	r.Use(gin.Recovery())
	r.GET("/health", c.handleHealth)
	return r
}

func (c *Core) handleHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"health": true})
}

func loggerMiddleware(c *gin.Context) {
	path := c.Request.URL.Path
	start := time.Now()
	c.Next()
	latency := time.Since(start)
	log.Printf("%3d | %15s | %s %s %10v \"%s\"%s\n",
		c.Writer.Status(),
		c.ClientIP(),
		c.Request.Method,
		path,
		latency,
		c.Request.UserAgent(),
		c.Errors.ByType(gin.ErrorTypePrivate).String(),
	)
}

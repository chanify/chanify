package core

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/chanify/chanify/logic"
	"github.com/gin-gonic/gin"
)

// error define
var (
	ErrNoContent       = errors.New("NoContent")
	ErrTooLargeContent = errors.New("TooLargeContent")
	ErrInvalidContent  = errors.New("InvalidContent")
)

// Core instance
type Core struct {
	mcpMu       sync.Mutex
	mcpStopped  bool
	mcpRequests sync.WaitGroup
	admin       *adminState
	logic       *logic.Logic
	mcpHandler  http.Handler
	mcpContext  context.Context
	mcpCancel   context.CancelFunc
}

// New core instance
func New() *Core {
	gin.SetMode(gin.ReleaseMode)
	return &Core{}
}

// Init core with option
func (c *Core) Init(opts *logic.Options) error {
	hash, err := adminPasswordHash(opts.AdminPassword)
	if err != nil {
		return err
	}
	c.logic, err = logic.NewLogic(opts)
	if err == nil && opts.MCPEnabled {
		c.mcpContext, c.mcpCancel = context.WithCancel(context.Background())
		c.mcpHandler = c.initMCP(opts.MCPListen)
	}
	if err == nil && hash != "" {
		if !c.logic.AdminAvailable() {
			c.Close()
			return errors.New("web administration requires a persistent node")
		}
		c.admin = newAdmin(hash, c.logic.Endpoint)
	}
	return err
}

// Close & cleaup for core
func (c *Core) Close() {
	c.mcpMu.Lock()
	c.mcpStopped = true
	if c.mcpCancel != nil {
		c.mcpCancel()
	}
	c.mcpMu.Unlock()
	c.mcpRequests.Wait()
	if c.logic != nil {
		c.logic.Close()
		c.logic = nil
	}
}

// APIHandler return handler for http
func (c *Core) APIHandler() http.Handler {
	r := gin.New()
	r.SetTrustedProxies([]string{"0.0.0.0/0", "::/0"}) // nolint: errcheck
	r.Use(loggerMiddleware)
	r.Use(gin.Recovery())
	r.GET("/", c.handleHome)
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"health": true})
	})
	r.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"res": http.StatusNotFound, "msg": "not found"})
	})

	if c.mcpHandler != nil {
		r.Any("/mcp", gin.WrapH(c.mcpHandler))
	}
	if c.admin != nil {
		handler := c.adminHandler()
		r.Any("/admin", gin.WrapH(handler))
		r.Any("/admin/*path", gin.WrapH(handler))
	}
	s := r.Group("/v1")
	s.GET("/sender/:token/:msg", c.handleSender)
	s.GET("/sender/:token/", c.handleSender)
	s.GET("/sender/:token", c.handleSender)
	s.POST("/sender/*token", c.handlePostSender)
	s.POST("/sender", c.handlePostSender)
	s.POST("/webhook/:name/:token", c.handlePostWebhook)
	s.POST("/webhook/:name", c.handlePostWebhook)

	api := r.Group("/rest/v1")
	api.GET("/info", c.handleInfo)
	api.GET("/qrcode", c.handleQRCode)
	api.POST("/bind-user", c.handleBindUser)
	api.POST("/unbind-user", c.handleUnbindUser)
	api.POST("/delete-user", c.handleDeleteUser)
	api.POST("/server-tokens", c.handleServerTokens)
	api.POST("/mcp-tokens", c.handleMCPTokens)
	api.POST("/push-token", c.handleUpdatePushToken)

	file := r.Group("/files")
	file.GET("/images/:fname", c.handleImageDownload)
	file.GET("/audios/:fname", c.handleAudioDownload)
	file.GET("/files/:fname", c.handleFileDownload)

	return r
}

func (c *Core) handleHome(ctx *gin.Context) {
	c.handleQRCode(ctx)
}

func (c *Core) handleInfo(ctx *gin.Context) {
	info, sign := c.logic.GetInfo()
	ctx.DataFromReader(http.StatusOK, int64(len(info)), "application/json; charset=utf-8", bytes.NewReader(info), map[string]string{
		"CHSign-Node": sign,
	})
}

func (c *Core) handleQRCode(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "image/png", c.logic.GetQRCode())
}

func loggerMiddleware(c *gin.Context) {
	path := c.Request.URL.Path
	start := time.Now()
	c.Next()
	latency := time.Since(start)
	if len(path) > 64 {
		path = path[:64]
	}
	log.Printf("%3d | %15s | %s %s %10v \"%s\"%s\n",
		c.Writer.Status(),
		fixLog(c.ClientIP()),
		c.Request.Method,
		fixLog(path),
		latency,
		fixLog(c.Request.UserAgent()),
		c.Errors.ByType(gin.ErrorTypePrivate).String(),
	)
}

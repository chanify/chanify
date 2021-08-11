package core

import (
	"bytes"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
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
	logic *logic.Logic
}

// New core instance
func New() *Core {
	gin.SetMode(gin.ReleaseMode)
	return &Core{}
}

// Init core with option
func (c *Core) Init(opts *logic.Options) error {
	var err error
	c.logic, err = logic.NewLogic(opts)
	return err
}

// Close & cleaup for core
func (c *Core) Close() {
	if c.logic != nil {
		c.logic.Close()
		c.logic = nil
	}
}

// APIHandler return handler for http
func (c *Core) APIHandler() http.Handler {
	r := gin.New()
	r.Use(loggerMiddleware)
	r.Use(gin.Recovery())
	r.GET("/", c.handleHome)
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"health": true})
	})
	r.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"res": http.StatusNotFound, "msg": "not found"})
	})

	s := r.Group("/v1")
	s.GET("/sender/:token/:msg", c.handleSender)
	s.GET("/sender/:token/", c.handleSender)
	s.POST("/sender/*token", c.handlePostSender)
	s.POST("/sender", c.handlePostSender)

	s = r.Group("/v2")
	s.GET("/sender/:uid/:msg", c.handleUserSender)

	api := r.Group("/rest/v1")
	api.GET("/info", c.handleInfo)
	api.GET("/qrcode", c.handleQRCode)
	api.POST("/bind-user", c.handleBindUser)
	api.POST("/unbind-user", c.handleUnbindUser)
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
		fixClientIP(c),
		c.Request.Method,
		path,
		latency,
		c.Request.UserAgent(),
		c.Errors.ByType(gin.ErrorTypePrivate).String(),
	)
}

var remoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}

func fixClientIP(c *gin.Context) string {
	// ref: https://github.com/gin-gonic/gin/issues/2697
	for _, key := range remoteIPHeaders {
		realIP, valid := validateHeader(c.GetHeader(key))
		if valid {
			return realIP
		}
	}
	return c.ClientIP()
}

func validateHeader(header string) (clientIP string, valid bool) {
	if header == "" {
		return "", false
	}
	items := strings.Split(header, ",")
	for i, ipStr := range items {
		ipStr = strings.TrimSpace(ipStr)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return "", false
		}

		// We need to return the first IP in the list, but,
		// we should not early return since we need to validate that
		// the rest of the header is syntactically valid
		if i == 0 {
			clientIP = ipStr
			valid = true
		}
	}
	return
}

package core

import (
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/chanify/chanify/logic"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var (
	base64Encode = base64.RawURLEncoding
)

type Core struct {
	logic *logic.Logic
}

func New() *Core {
	gin.SetMode(gin.ReleaseMode)
	return &Core{}
}

func (c *Core) Init(opts *logic.Options) error {
	var err error
	c.logic, err = logic.NewLogic(opts)
	return err
}

func (c *Core) Close() {
	if c.logic != nil {
		c.logic.Close()
		c.logic = nil
	}
}

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

	api := r.Group("/rest/v1")
	api.GET("/info", c.handleInfo)
	api.GET("/qrcode", c.handleQRCode)
	api.POST("/bind-user", c.handleBindUser)
	api.POST("/unbind-user", c.handleUnbindUser)
	api.POST("/push-token", c.handleUpdatePushToken)
	return r
}

func (c *Core) handleHome(ctx *gin.Context) {
	c.handleQRCode(ctx)
}

func (c *Core) handleInfo(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c.logic.GetInfo())
}

func (c *Core) handleQRCode(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "image/png", c.logic.GetQRCode())
}

func (c *Core) handleUpdatePushToken(ctx *gin.Context) {
	var params struct {
		Nonce    uint64 `json:"nonce"`
		DeviceID string `json:"device"`
		UserID   string `json:"user"`
		Token    string `json:"token"`
		Sandbox  bool   `json:"sandbox,omitempty"`
	}
	if err := ctx.ShouldBindBodyWith(&params, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid params"})
		return
	}
	u, err := c.logic.GetUser(params.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user id"})
		return
	}
	if u.IsServerless() {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user mode"})
		return
	}
	if !ValidateUser(ctx, u.GetPublicKeyString()) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid user sign"})
		return
	}
	dev, err := c.logic.GetDeviceKey(params.DeviceID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid device id"})
		return
	}
	if !ValidateDevice(ctx, base64Encode.EncodeToString(dev)) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid device sign"})
		return
	}
	if err := c.logic.UpdatePushToken(params.UserID, params.DeviceID, params.Token, params.Sandbox); err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"res": http.StatusConflict, "msg": "update push token failed"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"uuid": params.DeviceID,
		"uid":  params.UserID,
	})
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
		c.ClientIP(),
		c.Request.Method,
		path,
		latency,
		c.Request.UserAgent(),
		c.Errors.ByType(gin.ErrorTypePrivate).String(),
	)
}

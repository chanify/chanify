package core

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Core struct {
	info   ServerInfo
	engine *gin.Engine
}

func New(name string, version string, endpoint string) *Core {
	gin.SetMode(gin.ReleaseMode)
	c := &Core{}
	c.info.Name = name
	c.info.Version = version
	c.setEndpoint(endpoint)
	return c
}

func (c *Core) Init(data string) error {
	// data, _ = homedir.Expand(data)
	// if _, err := os.Stat(data); os.IsNotExist(err) {
	// 	return errors.New("invalid data directory")
	// }
	return c.initFeatures()
}

func (c *Core) Close() {
}

func (c *Core) APIHandler() http.Handler {
	if c.engine == nil {
		r := gin.New()
		c.engine = r
		r.Use(loggerMiddleware)
		r.Use(gin.Recovery())
		r.GET("/", c.handleHome)
		r.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"health": true})
		})
		r.NoRoute(func(ctx *gin.Context) {
			ctx.JSON(http.StatusNotFound, gin.H{"res": http.StatusNotFound, "msg": "not found"})
		})

		api := r.Group("/rest/v1")
		api.GET("/info", c.handleInfo)
		api.GET("/qrcode", c.handleQrCode)
		api.POST("/bind-user", c.handleBindUser)
		api.POST("/sender", c.handleSender)
	}
	return c.engine
}

func (c *Core) handleHome(ctx *gin.Context) {
	ctx.Request.URL.Path = "/rest/v1/qrcode"
	c.engine.HandleContext(ctx)
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

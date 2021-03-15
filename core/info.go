package core

import (
	"crypto/sha256"
	"log"
	"net/http"
	"net/url"
	"os"

	cc "github.com/chanify/chanify/crypto"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/go-homedir"
	qrcode "github.com/skip2/go-qrcode"
)

type Feature struct {
	Code string `json:"code"`
}

type ServerInfo struct {
	NodeId    string   `json:"nodeid"`
	Name      string   `json:"name,omitempty"`
	Version   string   `json:"version"`
	PublicKey string   `json:"pubkey"`
	Endpoint  string   `json:"endpoint,omitempty"`
	Features  []string `json:"features,omitempty"`

	key    *cc.SecretKey `json:"-"`
	qrCode []byte        `json:"-"`
	secret []byte        `json:"-"`
}

func (c *Core) SetDataDir(datadir string) {
	datadir, _ = homedir.Expand(datadir)
	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		log.Println("data dir:", err)
	}
}

func (c *Core) SetSecret(secret string) {
	c.info.secret = sha256.New().Sum([]byte(secret))
	c.info.key, _ = cc.GenerateSecretKey([]byte(secret))
	c.info.PublicKey = c.info.key.EncodePublicKey()
	c.info.NodeId = c.info.key.ToID(0x01)
}

func (c *Core) SetVersion(ver string) {
	c.info.Version = ver
}

func (c *Core) SetEndpoint(endpoint string) {
	c.info.Endpoint = endpoint
	c.info.qrCode, _ = qrcode.Encode("chanify://node?endpoint="+url.QueryEscape(endpoint), qrcode.Medium, 256)
}

func (c *Core) SetName(name string) {
	c.info.Name = name
}

func (c *Core) InitFeatures() {
	c.info.Features = []string{"msg.text"}
}

func (c *Core) handleInfo(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c.info)
}

func (c *Core) handleQrCode(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "image/png", c.info.qrCode)
}

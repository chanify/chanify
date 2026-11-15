package core

import (
	"crypto/hmac"
	"crypto/sha256"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
	"github.com/chanify/chanify/pb"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
	"google.golang.org/protobuf/proto"
)

func TestSendingRejectsInvalidChannelAcrossInputs(t *testing.T) {
	c, _, _ := mcpFixture(t, true, false)
	key := crypto.GenerateSecretKey([]byte("synthetic-mcp-account"))
	uid := key.ToID(0)
	user, err := c.logic.GetUser(uid)
	if err != nil {
		t.Fatal(err)
	}
	data, err := proto.Marshal(&pb.Token{UserId: uid, NodeId: c.logic.NodeID, Scope: pb.TokenScope_ServerSend, TokenId: make([]byte, 16), Expires: uint64(time.Now().Add(time.Hour).Unix())})
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, user.SecretKey[:32])
	_, _ = mac.Write(data)
	raw := "cst1:" + crypto.Base64Encode.EncodeToString(data) + ".." + crypto.Base64Encode.EncodeToString(mac.Sum(nil))
	if _, err := c.logic.ManageServerTokens(uid, "create", "", "Test", raw); err != nil {
		t.Fatal(err)
	}
	tk, err := model.ParseToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"image", "audio", "get", "lua"} {
		t.Run(kind, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("GET", "/v1/sender?channel=invalid!&text=hello", nil)
			ctx.Request.Header.Set("token", raw)
			switch kind {
			case "image":
				if _, err := (&MsgParam{Token: tk, Channel: "invalid!", ChannelProvided: true}).ParseImage(c, ctx); err == nil || !strings.Contains(err.Error(), "invalid channel") {
					t.Fatal("invalid image channel accepted", err)
				}
			case "audio":
				if _, err := (&MsgParam{Token: tk, Channel: "invalid!", ChannelProvided: true}).ParseAudio(c, ctx); err == nil || !strings.Contains(err.Error(), "invalid channel") {
					t.Fatal("invalid audio channel accepted", err)
				}
			case "get":
				c.handleSender(ctx)
				if w.Code != 400 {
					t.Fatal(w.Code)
				}
			case "lua":
				l := lua.NewState()
				defer l.Close()
				ctx.Set(coreKey, c)
				initHttpLua(l, ctx)
				if err := l.DoString(`return ctx:send({text="hello",channel="invalid!"})`); err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(l.Get(-1).String(), `"res":400`) {
					t.Fatal(l.Get(-1))
				}
			}
		})
	}
}

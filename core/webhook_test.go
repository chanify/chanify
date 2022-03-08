package core

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chanify/chanify/logic"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

func TestWebHook(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "plugin")
	whdir := filepath.Join(dir, "webhook")
	os.MkdirAll(whdir, os.ModePerm) // nolint: errcheck
	fpath := filepath.Join(whdir, "github.lua")
	fs, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	defer fs.Close()
	fs.WriteString("ctx:token();ctx:body();ctx:header(\"\");return 201") // nolint: errcheck
	fs.Sync()                                                            // nolint: errcheck

	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", PluginPath: dir}) // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/v1/webhook/github", nil)
	c.APIHandler().ServeHTTP(w, ctx.Request)
	if w.Result().StatusCode != 201 {
		t.Fatal("Do webhook failed")
	}
}

func TestWebHookFailed(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "plugin")
	whdir := filepath.Join(dir, "webhook")
	os.MkdirAll(whdir, os.ModePerm) // nolint: errcheck
	fpath := filepath.Join(whdir, "github.lua")
	fs, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	defer fs.Close()
	fs.WriteString("a()") // nolint: errcheck
	fs.Sync()             // nolint: errcheck

	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", PluginPath: dir}) // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/v1/webhook/github", strings.NewReader(`{}`))
	c.APIHandler().ServeHTTP(w, ctx.Request)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("Check do webhook failed")
	}
}

func TestWebHookNotFound(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/v1/webhook/test", nil)
	c.handlePostWebhook(ctx)
	if w.Result().StatusCode != http.StatusNotFound {
		t.Fatal("Check webhook failed")
	}
}

func TestGetHttpLuaNoReturn(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	if err := l.DoString(``); err != nil {
		t.Fatal(err)
	}
	c, ct, d := getHttpLuaReturn(l)
	if c != 200 || ct != "text/plain; charset=utf-8" || len(d) != 0 {
		t.Error("Return value 1 failed:", c, ct, d)
	}
}

func TestGetHttpLuaReturn1(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	if err := l.DoString(`return 401`); err != nil {
		t.Fatal(err)
	}
	c, ct, d := getHttpLuaReturn(l)
	if c != 401 || ct != "text/plain; charset=utf-8" || len(d) != 0 {
		t.Error("Return value 1 failed:", c, ct, d)
	}
}

func TestGetHttpLuaReturn2(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	if err := l.DoString(`return 201, "abc"`); err != nil {
		t.Fatal(err)
	}
	c, ct, d := getHttpLuaReturn(l)
	if c != 201 || ct != "text/plain; charset=utf-8" || d != "abc" {
		t.Error("Return value 2 failed:", c, ct, d)
	}
}

func TestGetHttpLuaReturn3(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	if err := l.DoString(`return 201, "application/json; charset=utf-8", "{}"`); err != nil {
		t.Fatal(err)
	}
	c, ct, d := getHttpLuaReturn(l)
	if c != 201 || ct != "application/json; charset=utf-8" || d != "{}" {
		t.Error("Return value 2 failed:", c, ct, d)
	}
}

func TestLuaCheckContext(t *testing.T) {
	l := lua.NewState()
	defer l.Close()

	mt := l.NewTypeMetatable("Context")
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), luaContextMethods))

	lc := l.NewUserData()
	lc.Value = nil
	lc.Metatable = mt
	l.SetGlobal("ctx", lc)
	if err := l.DoString(`ctx:token()`); err == nil {
		t.Error("Check context failed")
	}
}

func TestLuaCheckHttpBody(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set(gin.BodyBytesKey, []byte("{}"))
	initHttpLua(l, ctx)
	if err := l.DoString(`ctx:body()`); err != nil {
		t.Fatal(err)
	}
}

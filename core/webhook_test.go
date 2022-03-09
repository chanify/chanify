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
	s := `
		local req=ctx:request()
		req:token()
		req:body()
		req:header("")
		req:header("host")
		req:header("user-agent")
		req:header("content-length")
		assert(string.len(req:url()) > 0, "url error")
		assert(req:query("xyz") == nil, "query error")
		assert(req:query("abc") == "123", "query error")
		assert(ctx:env("z") == nil, "env error")
		return 201,ctx:env("x")`
	fs.WriteString(s) // nolint: errcheck
	fs.Sync()         // nolint: errcheck

	c := New()
	defer c.Close()
	whs := []map[string]interface{}{
		{"name": "github", "file": ""},
		{"name": "github", "file": fpath, "env": map[string]interface{}{"x": "123", "y": 456}},
		{"name": "github", "file": "x.lua"},
	}
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", PluginPath: dir, WebHooks: whs}) // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/v1/webhook/github?abc=123", nil)
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
	whs := []map[string]interface{}{
		{"name": "github", "file": fpath},
	}
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", PluginPath: dir, WebHooks: whs}) // nolint: errcheck
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
	if err := l.DoString(`ctx:request():token()`); err == nil {
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
	if err := l.DoString(`ctx:request():body()`); err != nil {
		t.Fatal(err)
	}
}

func TestLuaContextSend(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "nosql://?secret=123"}) // nolint: errcheck
	_, err := c.parseToken("CNjo6ua-WhIiQUJPTzZUU0lYS1NFVklKS1hMRFFTVVhRUlhVQU9YR0dZWQ..faqRNWqzTW3Fjg4xh9CS_p8IItEHjSQiYzJjxcqf_tg")
	if err != nil {
		t.Error(err)
	}

	l := lua.NewState()
	defer l.Close()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/v1/webhook/github?token=CNjo6ua-WhIiQUJPTzZUU0lYS1NFVklKS1hMRFFTVVhRUlhVQU9YR0dZWQ..faqRNWqzTW3Fjg4xh9CS_p8IItEHjSQiYzJjxcqf_tg", nil)
	ctx.Set(gin.BodyBytesKey, []byte("{}"))
	ctx.Set(coreKey, c)
	initHttpLua(l, ctx)

	if err := l.DoString(`ctx:send("")`); err != nil {
		t.Fatal(err)
	}
	if err := l.DoString(`ctx:send("abc")`); err != nil {
		t.Fatal(err)
	}
	if err := l.DoString(`ctx:send({text="123",sound=1})`); err != nil {
		t.Fatal(err)
	}
	if err := l.DoString(`ctx:send({text="123",token="abc"})`); err != nil {
		t.Fatal(err)
	}
	if err := l.DoString(`ctx:send({text="` + strings.Repeat("A", 4000) + `"})`); err != nil {
		t.Fatal(err)
	}
}

func TestLuaContextSendFailed(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set(gin.BodyBytesKey, []byte("{}"))
	initHttpLua(l, ctx)
	if err := l.DoString(`ctx:send({})`); err != nil {
		t.Fatal(err)
	}
}

func TestLuaSendContext(t *testing.T) {
	ctx := &luaSendContext{}
	ctx.DataFromReader(0, 3, "", strings.NewReader("abc"), nil)
	if ctx.msg != "abc" {
		t.Error("Check send context failed")
	}
}

func TestLuaGetOptsArray(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	tbl := l.NewTable()
	arr := l.NewTable()
	tbl.RawSetString("a", arr)
	arr.Append(lua.LNumber(10))
	arr.Append(lua.LNumber(20))
	ss := luaGetOptsArray(tbl, "a")
	if len(ss) != 2 && ss[1] == "20" {
		t.Error("Check opts array failed")
	}
}

package core

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
	"github.com/sideshow/apns2"
)

func TestSender(t *testing.T) {
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "nosql://?secret=123"}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("GET", "/v1/sender/CNjo6ua-WhIiQUJPTzZUU0lYS1NFVklKS1hMRFFTVVhRUlhVQU9YR0dZWQ..faqRNWqzTW3Fjg4xh9CS_p8IItEHjSQiYzJjxcqf_tg/123", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatal("Sender message failed")
	}
}

func TestSenderFailed(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "nosql://?secret=123"}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("GET", "/v1/sender/123/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatal("Check sender unauthorized failed")
	}

	req = httptest.NewRequest("GET", "/v1/sender/CNjo6ua-WhIiQUJPTzZUU0lYS1NFVklKS1hMRFFTVVhRUlhVQU9YR0dZWQ..faqRNWqzTW3Fjg4xh9CS_p8IItEHjSQiYzJjxcqf_tg/", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("Check sender failed")
	}
}

func TestSenderNull(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "nosql://?secret=123"}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/v1/sender", nil)
	req.Header.Set("Token", "CNjo6ua-WhIiQUJPTzZUU0lYS1NFVklKS1hMRFFTVVhRUlhVQU9YR0dZWQ..faqRNWqzTW3Fjg4xh9CS_p8IItEHjSQiYzJjxcqf_tg")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("Check sender null failed")
	}
}

func TestSenderPost(t *testing.T) {
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "nosql://?secret=123"}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/v1/sender", bytes.NewReader([]byte("Hello")))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Token", "CNjo6ua-WhIiQUJPTzZUU0lYS1NFVklKS1hMRFFTVVhRUlhVQU9YR0dZWQ..faqRNWqzTW3Fjg4xh9CS_p8IItEHjSQiYzJjxcqf_tg")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatal("Check send post failed")
	}
}

func TestSenderPostFailed(t *testing.T) {
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "nosql://?secret=123"}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/v1/sender", bytes.NewReader([]byte("Hello")))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatal("Check send post token failed")
	}
}

func TestSenderPostForm(t *testing.T) {
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "nosql://?secret=123"}) // nolint: errcheck
	handler := c.APIHandler()

	data := url.Values{
		"text":  {"123"},
		"token": {"CNjo6ua-WhIiQUJPTzZUU0lYS1NFVklKS1hMRFFTVVhRUlhVQU9YR0dZWQ..faqRNWqzTW3Fjg4xh9CS_p8IItEHjSQiYzJjxcqf_tg"},
	}
	req := httptest.NewRequest("POST", "/v1/sender", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatal("Check send post failed")
	}
}

func TestSenderPostFormData(t *testing.T) {
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "nosql://?secret=123"}) // nolint: errcheck
	handler := c.APIHandler()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	partText, _ := writer.CreateFormField("text")                                                                                      // nolint: errcheck
	partText.Write([]byte("hello"))                                                                                                    // nolint: errcheck                                                                                            // nolint: errcheck
	partTitle, _ := writer.CreateFormField("title")                                                                                    // nolint: errcheck
	partTitle.Write([]byte("MyTitle"))                                                                                                 // nolint: errcheck
	partSound, _ := writer.CreateFormField("sound")                                                                                    // nolint: errcheck
	partSound.Write([]byte("false"))                                                                                                   // nolint: errcheck                                                                                           // nolint: errcheck
	partPriority, _ := writer.CreateFormField("priority")                                                                              // nolint: errcheck
	partPriority.Write([]byte("5"))                                                                                                    // nolint: errcheck                                                                                           // nolint: errcheck
	partToken, _ := writer.CreateFormField("token")                                                                                    // nolint: errcheck
	partToken.Write([]byte("CNjo6ua-WhIiQUJPTzZUU0lYS1NFVklKS1hMRFFTVVhRUlhVQU9YR0dZWQ..faqRNWqzTW3Fjg4xh9CS_p8IItEHjSQiYzJjxcqf_tg")) // nolint: errcheck
	writer.Close()

	req := httptest.NewRequest("POST", "/v1/sender", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatal("Check send post form failed")
	}
}

func TestSenderPostJSON(t *testing.T) {
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "nosql://?secret=123"}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/v1/sender", strings.NewReader(`{
		"sound": 1,
		"title": "abc",
		"text": "hello",
		"token": "CNjo6ua-WhIiQUJPTzZUU0lYS1NFVklKS1hMRFFTVVhRUlhVQU9YR0dZWQ..faqRNWqzTW3Fjg4xh9CS_p8IItEHjSQiYzJjxcqf_tg"
	}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatal("Send post json failed")
	}
}

type MockAPNSPusher struct{}

func (m *MockAPNSPusher) Push(n *apns2.Notification) (*apns2.Response, error) {
	return nil, nil
}

func TestSendDirect(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	w := httptest.NewRecorder()
	tk, _ := model.ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRw..c2lnbg") // nolint: errcheck
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "", nil)
	c.SendDirect(ctx, tk, model.NewMessage(tk))
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("Check invalid user key failed")
	}

	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "", nil)
	c.logic.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", false)                                      // nolint: errcheck                                  // nolint: errcheck
	c.logic.BindDevice("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "B3BC1B875EDA13986801B1004B4ABF5760C197F4", "BDuFNLkmxyK0-NN3H3oKzzOtISq1w17-JAibD7X4pljYl6IEaEglWkKD5Iw537h-DYxAooXkHtu6un078sm7IiQ") // nolint: errcheck
	c.logic.UpdatePushToken("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "B3BC1B875EDA13986801B1004B4ABF5760C197F4", "aGVsbG8", false)                                                                     // nolint: errcheck                                            // nolint: errcheck
	c.logic.GetDevices("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY")                                                                                                                                        // nolint: errcheck
	logic.MockPusher = &MockAPNSPusher{}
	c.SendDirect(ctx, tk, model.NewMessage(tk).SetPriority(5))
	if w.Result().StatusCode != http.StatusOK {
		t.Fatal("Send direct failed")
	}
}

func TestSendForward(t *testing.T) {
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	w := httptest.NewRecorder()
	tk, _ := model.ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRw..c2lnbg") // nolint: errcheck
	msg := model.NewMessage(tk).DisableToken()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "", nil)
	c.SendForward(ctx, tk, msg)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("Check invalid user failed")
	}

	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "", nil)
	c.logic.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", true) // nolint: errcheck
	c.SendForward(ctx, tk, msg)
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Fatal("Check invalid key failed")
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusOK)
	}))
	defer ts.Close()
	logic.ApiEndpoint = ts.URL

	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "", nil)
	c.logic.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", true) // nolint: errcheck
	c.SendForward(ctx, tk, msg)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatal("Check invalid key failed")
	}
}

func TestSendMsg(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	w := httptest.NewRecorder()
	tk, _ := model.ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRw..c2lnbg") // nolint: errcheck
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "", nil)
	c.sendMsg(ctx, tk, model.NewMessage(tk).TextContent("123", "title").SetPriority(5))
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("Check invalid user failed")
	}

	logic.ApiEndpoint = "http://127.0.0.1"
	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "", nil)
	c.logic.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", true) // nolint: errcheck
	c.sendMsg(ctx, tk, model.NewMessage(tk).TextContent("123", "title"))
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Fatal("Check send serverless failed")
	}

	logic.MockPusher = &MockAPNSPusher{}
	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "", nil)
	c.logic.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", false) // nolint: errcheck
	c.sendMsg(ctx, tk, model.NewMessage(tk).TextContent("123", "title").SetPriority(5))
	if w.Result().StatusCode != http.StatusNotFound {
		t.Fatal("Check send serverful failed")
	}
}

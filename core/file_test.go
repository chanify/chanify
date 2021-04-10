package core

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

func TestImageFile(t *testing.T) {
	fpath := filepath.Join(os.TempDir(), "files")
	defer os.RemoveAll(fpath)
	os.MkdirAll(fpath+"/images/", os.ModePerm)                     // nolint: errcheck
	os.WriteFile(fpath+"/images/123456789", []byte("hello"), 0644) // nolint: errcheck
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", FilePath: fpath}) // nolint: errcheck

	tk, _ := model.ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRzIUx5tXg-Vym58og7aZw05IkoDvse8..c2lnbg") // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/files/images/123456789", nil)
	ctx.Request.URL.Path = "/files/images/123456789"
	ctx.Params = []gin.Param{{Key: "fname", Value: "123456789"}}
	c.downloadImageFile(ctx, tk)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatal("Download image hash failed", w.Result().StatusCode)
	}
}

func TestImageFileFailed(t *testing.T) {
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/files/images/123456789", nil)
	c.handleImageDownload(ctx)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatal("Check download token failed")
	}

	tk, _ := model.ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRw..c2lnbg") // nolint: errcheck
	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/files/images/123456789", nil)
	c.downloadImageFile(ctx, tk)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatal("Check download url hash failed")
	}

	tk, _ = model.ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRzIUx5tXg-Vym58og7aZw05IkoDvse8..c2lnbg") // nolint: errcheck
	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/files/images/123456789", nil)
	ctx.Request.URL.Path = "/files/images/123456789"
	c.downloadImageFile(ctx, tk)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("Check download image url hash failed")
	}

	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/files/images/123456789", nil)
	ctx.Request.URL.Path = "/files/images/123456789"
	ctx.Params = []gin.Param{{Key: "fname", Value: "123456789"}}
	c.downloadImageFile(ctx, tk)
	if w.Result().StatusCode != http.StatusNotFound {
		t.Fatal("Check download image name failed")
	}
}

func TestFile(t *testing.T) {
	fpath := filepath.Join(os.TempDir(), "files")
	defer os.RemoveAll(fpath)
	os.MkdirAll(fpath+"/files/", os.ModePerm)                     // nolint: errcheck
	os.WriteFile(fpath+"/files/123456789", []byte("hello"), 0644) // nolint: errcheck
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", FilePath: fpath}) // nolint: errcheck

	tk, _ := model.ParseToken("EgMxMjMiBGNoYW4qBU1GUkdHMhT9W_fNj-BHJX9yn6tO3jTtHjKyTA..c2lnbg") // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/files/files/123456789", nil)
	ctx.Request.URL.Path = "/files/files/123456789"
	ctx.Params = []gin.Param{{Key: "fname", Value: "123456789"}}
	c.downloadFile(ctx, tk)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatal("Download file hash failed", w.Result().StatusCode)
	}
}

func TestFileFailed(t *testing.T) {
	logic.ApiEndpoint = "http://127.0.0.1"
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/files/files/123456789", nil)
	c.handleFileDownload(ctx)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatal("Check download token failed")
	}

	tk, _ := model.ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRw..c2lnbg") // nolint: errcheck
	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/files/files/123456789", nil)
	c.downloadFile(ctx, tk)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatal("Check download url hash failed")
	}

	tk, _ = model.ParseToken("EgMxMjMiBGNoYW4qBU1GUkdHMhT9W_fNj-BHJX9yn6tO3jTtHjKyTA..c2lnbg") // nolint: errcheck
	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/files/files/123456789", nil)
	ctx.Request.URL.Path = "/files/files/123456789"
	c.downloadFile(ctx, tk)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("Check download file url hash failed")
	}

	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/files/files/123456789", nil)
	ctx.Request.URL.Path = "/files/files/123456789"
	ctx.Params = []gin.Param{{Key: "fname", Value: "123456789"}}
	c.downloadFile(ctx, tk)
	if w.Result().StatusCode != http.StatusNotFound {
		t.Fatal("Check download file name failed")
	}
}

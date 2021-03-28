package logic

import (
	"io"
	"testing"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
)

func TestLogic(t *testing.T) {
	l, err := NewLogic(&Options{
		DBUrl:  "sqlite://?mode=memory",
		Secret: "123",
	})
	if err != nil {
		t.Fatal("New logic failed:", err)
	}
	defer l.Close()
}

func TestLogicServerless(t *testing.T) {
	l, err := NewLogic(&Options{Secret: "123"})
	if err != nil {
		t.Fatal("New logic failed:", err)
	}
	defer l.Close()
}

func TestLogicFailed(t *testing.T) {
	if _, err := NewLogic(&Options{DBUrl: "nodriver://"}); err == nil {
		t.Fatal("Check logic dburl failed")
	}
	if _, err := NewLogic(&Options{}); err == nil {
		t.Fatal("Check logic secret failed")
	}
}

func TestUser(t *testing.T) {
	l, _ := NewLogic(&Options{Secret: "123"})
	if _, err := l.GetUserKey("GEZDG"); err != nil {
		t.Fatal("Get user key failed")
	}
	if _, err := l.GetUserKey("123"); err == nil {
		t.Fatal("Check get user key failed")
	}
}

func TestSaveImageFileFailed(t *testing.T) {
	l, _ := NewLogic(&Options{DBUrl: "sqlite://?mode=memory"})
	l.filepath = " "
	if _, err := l.SaveImageFile(nil); err != ErrInvalidContent {
		t.Fatal("Check image data failed")
	}
	if _, err := l.SaveImageFile([]byte("123")); err == nil {
		t.Fatal("Check image save failed")
	}
}

func TestGetAPNS(t *testing.T) {
	l, _ := NewLogic(&Options{DBUrl: "sqlite://?mode=memory"})
	MockPusher = nil
	if l.GetAPNS(false) != l.apnsPClient {
		t.Error("Get product apns failed")
	}
	if l.GetAPNS(true) != l.apnsDClient {
		t.Error("Get sandbox apns failed")
	}
}

func TestFixDataPath(t *testing.T) {
	opts := &Options{DataPath: "/"}
	opts.fixOptions()
	if len(opts.DBUrl) <= 0 {
		t.Error("Fix data path failed")
	}
}

func TestFixSecretKey(t *testing.T) {
	l, _ := NewLogic(&Options{DBUrl: "nosql://?secret=123456"})
	if err := l.fixSecretKey(); err != nil {
		t.Fatal("Fix secret key failed:", err)
	}
	l.secKey = nil
	if err := l.fixSecretKey(); err == nil {
		t.Fatal("Check fix secret key failed")
	}
}

func TestUpsertUserFailed(t *testing.T) {
	l, _ := NewLogic(&Options{DBUrl: "nosql://?secret=123456"})
	if _, err := l.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", false); err == nil {
		t.Fatal("Check upsert user serverful failed")
	}
	if _, err := l.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", true); err != nil {
		t.Fatal("Upsert user serverless failed")
	}
	if _, err := l.createUser("123", crypto.GenerateSecretKey([]byte{}).GetPublicKey(), true); err == nil {
		t.Fatal("Check create user key failed")
	}

	oldReader := randReader
	defer func() {
		randReader = oldReader
	}()
	randReader = func(b []byte) (n int, err error) {
		return 0, io.EOF
	}
	l, _ = NewLogic(&Options{DBUrl: "sqlite://?mode=memory"})
	if _, err := l.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", false); err == nil {
		t.Fatal("Check upsert user serverless failed")
	}
}

func TestVerifyToken(t *testing.T) {
	l, _ := NewLogic(&Options{DBUrl: "sqlite://?mode=memory"})
	tk, _ := model.ParseToken("CNjo6ua-WhIiQUJPTzZUU0lYS1NFVklKS1hMRFFTVVhRUlhVQU9YR0dZWQ..faqRNWqzTW3Fjg4xh9CS_p8IItEHjSQiYzJjxcqf_tg")
	if l.VerifyToken(tk) {
		t.Fatal("Check invalid user token failed")
	}
}

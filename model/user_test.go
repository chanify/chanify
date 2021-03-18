package model

import "testing"

func TestUser(t *testing.T) {
	u := &User{}
	u.SetServerless(false)
	if u.IsServerless() {
		t.Fatal("Set serverless false failed")
	}
	u.SetServerless(true)
	if !u.IsServerless() {
		t.Fatal("Set serverless true failed")
	}
}

func TestCalcUser(t *testing.T) {
	_, err := CalcUserKey("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4")
	if err != nil {
		t.Error("CalcUserKey failed:", err)
	}
	if _, err := CalcUserKey("", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"); err == nil {
		t.Error("Check user id failed:", err)
	}
	if _, err := CalcUserKey("", "***"); err == nil {
		t.Error("Check key format failed")
	}
	if _, err := CalcUserKey("", "aGVsbG8"); err == nil {
		t.Error("Check user key failed")
	}
}

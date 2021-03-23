package model

import "testing"

func TestNewAESGCM(t *testing.T) {
	if _, err := NewAESGCM([]byte{}); err == nil {
		t.Error("Check new aes gcm failed")
	}
}

func TestDecodePushToken(t *testing.T) {
	d, err := DecodePushToken("aGVsbG8")
	if err != nil {
		t.Fatal("Decode push token failed:", err)
	}
	if string(d) != "hello" {
		t.Fatal("Decode push token value failed:", err)
	}
}

func TestCalcDeviceKey(t *testing.T) {
	_, err := CalcDeviceKey("B3BC1B875EDA13986801B1004B4ABF5760C197F4", "BDuFNLkmxyK0-NN3H3oKzzOtISq1w17-JAibD7X4pljYl6IEaEglWkKD5Iw537h-DYxAooXkHtu6un078sm7IiQ")
	if err != nil {
		t.Fatal("CalcDeviceKey failed:", err)
	}
	if _, err := CalcDeviceKey("B3BC1B875EDA13986801B1004B4ABF5760C197F4", "AAAFNLkmxyK0-NN3H3oKzzOtISq1w17-JAibD7X4pljYl6IEaEglWkKD5Iw537h-DYxAooXkHtu6un078sm7IiQ"); err == nil {
		t.Fatal("Check device key format failed")
	}
	if _, err := CalcDeviceKey("***", "BDuFNLkmxyK0-NN3H3oKzzOtISq1w17-JAibD7X4pljYl6IEaEglWkKD5Iw537h-DYxAooXkHtu6un078sm7IiQ"); err == nil {
		t.Fatal("Check device uuid failed")
	}
	if _, err := CalcDeviceKey("B3BC1B875EDA13986801B1004B4ABF5760C197F4", "****"); err == nil {
		t.Fatal("Check device key failed")
	}

}

func TestCalcUser(t *testing.T) {
	_, err := CalcUserKey("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4")
	if err != nil {
		t.Fatal("CalcUserKey failed:", err)
	}
	if _, err := CalcUserKey("", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"); err == nil {
		t.Fatal("Check user id failed:", err)
	}
	if _, err := CalcUserKey("", "***"); err == nil {
		t.Fatal("Check key format failed")
	}
	if _, err := CalcUserKey("", "aGVsbG8"); err == nil {
		t.Fatal("Check user key failed")
	}
}

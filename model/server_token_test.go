package model

import (
	"crypto/hmac"
	"crypto/sha256"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

func signedServerToken(t *testing.T, data *pb.Token) string {
	t.Helper()
	raw, err := proto.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, make([]byte, 32))
	_, _ = mac.Write(raw)
	return "cst1:" + crypto.Base64Encode.EncodeToString(raw) + ".." + crypto.Base64Encode.EncodeToString(mac.Sum(nil))
}
func TestServerToken(t *testing.T) {
	data := &pb.Token{UserId: "user", NodeId: "MFRGG", Expires: uint64(time.Now().Add(time.Hour).Unix()), Scope: pb.TokenScope_ServerSend, TokenId: make([]byte, 16)}
	raw := signedServerToken(t, data)
	tk, err := ParseToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !tk.VerifySign(make([]byte, 32)) || tk.VerifySign([]byte("short")) || tk.RawToken() != raw {
		t.Fatal("signature or raw token changed")
	}
	for _, channel := range []string{"", "invalid!", crypto.Base64Encode.EncodeToString([]byte("garbage")), "CAEQAw", "CAEQASoBYQ"} {
		if _, err := tk.ForChannel(channel); err == nil {
			t.Fatalf("accepted invalid channel %q", channel)
		}
	}

	defaulted, err := tk.ForOptionalChannel("", false)
	if err != nil || string(defaulted.GetChannel()) != string(defaultChannel) {
		t.Fatal("omitted destination did not select Uncategorized", err)
	}
	if _, err := tk.ForOptionalChannel("", true); err == nil {
		t.Fatal("explicit empty channel defaulted")
	}
	channel, _ := proto.Marshal(&pb.Channel{Type: pb.ChanType_User, Name: "development"})
	resolved, err := tk.ForChannel(crypto.Base64Encode.EncodeToString(channel))
	if err != nil {
		t.Fatal(err)
	}
	if string(resolved.GetChannel()) != string(channel) || len(tk.GetChannel()) != 0 || !resolved.VerifySign(make([]byte, 32)) || resolved.HashValue() == nil {
		t.Fatal("destination mutated signed credential")
	}
	if _, err := ParseToken(strings.TrimPrefix(raw, "cst1:")); err == nil {
		t.Fatal("accepted scope without prefix")
	}
	for _, mutate := range []func(*pb.Token){
		func(p *pb.Token) { p.Scope = pb.TokenScope_ChannelSend }, func(p *pb.Token) { p.Scope = 99 },
		func(p *pb.Token) { p.Channel = channel }, func(p *pb.Token) { p.TokenId = nil }, func(p *pb.Token) { p.NodeId = "" }, func(p *pb.Token) { p.DeviceId = []byte("device") },
	} {
		candidate := proto.Clone(data).(*pb.Token)
		mutate(candidate)
		if _, err := ParseToken(signedServerToken(t, candidate)); err == nil {
			t.Fatal("accepted malformed server token")
		}
	}
}

package model

import (
	"strings"
	"sync"
	"testing"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

func TestServerTokenRejectsInvalidChannelCode(t *testing.T) {
	tk := &Token{data: pb.Token{Scope: pb.TokenScope_ServerSend}}
	if _, err := tk.ForChannel(strings.Repeat("x", 1025)); err == nil {
		t.Fatal("oversized channel accepted")
	}
	for _, code := range []string{"", "sys.device", "sys.unknown", "white space", "-dev", "_dev", ".dev", "dev/ops", " dev", "dev ", "dev\n", "dev\x00", "开发", strings.Repeat("x", 65)} {
		if _, err := tk.ForChannel(code); err == nil {
			t.Fatal("invalid channel accepted", code)
		}
	}
}

func TestMCPTokenMissingRecord(t *testing.T) {
	db, m := mockStore(t)
	m.ExpectQuery("SELECT").WillReturnError(ErrInvalidToken)
	if err := updateMCPToken(db, "user", "id", "name", false); err != ErrInvalidToken {
		t.Fatal(err)
	}
}

func TestServerTokenReadableChannels(t *testing.T) {
	tk := &Token{data: pb.Token{Scope: pb.TokenScope_ServerSend}}
	for _, tc := range []struct {
		code   string
		kind   pb.ChanType
		system pb.ChanCode
		name   string
	}{
		{"dev", pb.ChanType_User, pb.ChanCode_Uncategorized, "dev"},
		{"a", pb.ChanType_User, pb.ChanCode_Uncategorized, "a"},
		{"0_A.b-9", pb.ChanType_User, pb.ChanCode_Uncategorized, "0_A.b-9"},
		{strings.Repeat("a", 64), pb.ChanType_User, pb.ChanCode_Uncategorized, strings.Repeat("a", 64)},
		{"sys.none", pb.ChanType_Sys, pb.ChanCode_Uncategorized, ""},
		{"sys.timesets", pb.ChanType_Sys, pb.ChanCode_TimeSets, ""},
		{"CAIaA2Rldg", pb.ChanType_User, pb.ChanCode_Uncategorized, "CAIaA2Rldg"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			resolved, err := tk.ForChannel(tc.code)
			if err != nil {
				t.Fatal(err)
			}
			var channel pb.Channel
			if err := proto.Unmarshal(resolved.GetChannel(), &channel); err != nil {
				t.Fatal(err)
			}
			if channel.Type != tc.kind || channel.Code != tc.system || channel.Name != tc.name {
				t.Fatalf("unexpected destination: %v", &channel)
			}
		})
	}
}

func TestChannelTokenRetainsSignedDestination(t *testing.T) {
	data, err := ChannelData("fixed")
	if err != nil {
		t.Fatal(err)
	}
	tk := &Token{data: pb.Token{Channel: data}}
	for _, code := range []string{"dev", "sys.none", "invalid!", ""} {
		resolved, err := tk.ForOptionalChannel(code, true)
		if err != nil || resolved != tk || string(resolved.GetChannel()) != string(data) {
			t.Fatalf("fixed destination changed for %q: %v", code, err)
		}
	}
	resolved, err := tk.ForOptionalChannel("", false)
	if err != nil || resolved != tk {
		t.Fatalf("omitted channel changed fixed token: %v", err)
	}
}

// A registered credential may be reused concurrently for different destinations.
func TestServerTokenConcurrentChannelIsolation(t *testing.T) {
	raw := signedServerToken(t, &pb.Token{UserId: "synthetic-user", NodeId: "MFRGG", Scope: pb.TokenScope_ServerSend, TokenId: make([]byte, 16)})
	tk, err := ParseToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for _, code := range []string{"dev", "production", "sys.none", "sys.timesets"} {
		workers.Add(1)
		go func(code string) {
			defer workers.Done()
			for i := 0; i < 20; i++ {
				resolved, err := tk.ForOptionalChannel(code, true)
				if err != nil {
					t.Error(err)
					return
				}
				var got pb.Channel
				if err := proto.Unmarshal(resolved.GetChannel(), &got); err != nil {
					t.Error(err)
					return
				}
				expected, err := ChannelData(code)
				if err != nil {
					t.Error(err)
					return
				}
				if string(resolved.GetChannel()) != string(expected) || resolved.RawToken() != raw || !resolved.VerifySign(make([]byte, 32)) {
					t.Errorf("destination or signature changed for %q: %v", code, &got)
					return
				}
			}
		}(code)
	}
	workers.Wait()
	if len(tk.GetChannel()) != 0 {
		t.Fatal("per-request destination mutated original credential")
	}
	// Omitting channel after explicit requests must still select Uncategorized.
	resolved, err := tk.ForOptionalChannel("", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(resolved.GetChannel()) != string(defaultChannel) {
		t.Fatal("previous request changed default destination")
	}
}

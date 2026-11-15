package model

import (
	"strings"
	"testing"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

func TestServerTokenRejectsInvalidChannelStructure(t *testing.T) {
	tk := &Token{data: pb.Token{Scope: pb.TokenScope_ServerSend}}
	if _, err := tk.ForChannel(strings.Repeat("x", 1025)); err == nil {
		t.Fatal("oversized channel accepted")
	}
	for _, ch := range []*pb.Channel{
		{Type: pb.ChanType_Sys, Code: pb.ChanCode_Device},
		{Type: pb.ChanType_Sys, Name: "name"},
		{Type: pb.ChanType_User, Name: " "},
		{Type: pb.ChanType_User, Name: strings.Repeat("x", 257)},
		{Type: pb.ChanType(99)},
	} {
		data, err := proto.Marshal(ch)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tk.ForChannel(crypto.Base64Encode.EncodeToString(data)); err == nil {
			t.Fatal("invalid channel accepted", ch)
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

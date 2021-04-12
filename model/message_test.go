package model

import (
	"testing"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

func TestImageContent(t *testing.T) {
	tk, _ := ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRzIUx5tXg-Vym58og7aZw05IkoDvse8..c2lnbg")
	m := NewMessage(tk)
	m.ImageContent("", NewThumbnail(10, 20), 10)
	var ctx pb.MsgContent
	if err := proto.Unmarshal(m.Content, &ctx); err != nil {
		t.Fatal("Unmarshal image content failed")
	}
	if ctx.Thumbnail == nil {
		t.Fatal("Unmarshal image thumbnail failed")
	}
	if ctx.Thumbnail.Width != 10 || ctx.Thumbnail.Height != 20 {
		t.Fatal("Check image thumbnail failed")
	}
}

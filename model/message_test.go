package model

import (
	"bytes"
	"testing"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

func TestInterruptionLevel(t *testing.T) {
	tk, _ := ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRzIUx5tXg-Vym58og7aZw05IkoDvse8..c2lnbg")
	m := NewMessage(tk)
	m.SetInterruptionLevel("passive")
	if m.InterruptionLevel != pb.InterruptionLevel_IlPassive {
		t.Error("Interruption level passive failed!")
	}
	m.SetInterruptionLevel("active")
	if m.InterruptionLevel != pb.InterruptionLevel_IlActive {
		t.Error("Interruption level active failed!")
	}
	m.SetInterruptionLevel("time-sensitive")
	if m.InterruptionLevel != pb.InterruptionLevel_IlTimeSensitive {
		t.Error("Interruption level time sensitive failed!")
	}
}

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

func TestActionContent(t *testing.T) {
	tk, _ := ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRzIUx5tXg-Vym58og7aZw05IkoDvse8..c2lnbg")
	m := NewMessage(tk)
	m.ActionContent("123", "abc", []string{"1|http://127.0.0.1", "2|http://127.0.0.1", "3|http://127.0.0.1", "4|http://127.0.0.1", "5|http://127.0.0.1"})
	var ctx pb.MsgContent
	if err := proto.Unmarshal(m.Content, &ctx); err != nil {
		t.Fatal("Unmarshal image content failed")
	}
	if ctx.Actions == nil {
		t.Fatal("Unmarshal actions failed")
	}
	if len(ctx.Actions) != 4 {
		t.Fatal("Check actions failed")
	}
}

func TestTimeContent(t *testing.T) {
	tk, _ := ParseToken("EiJBQk9PNlRTSVhLU0VWSUpLWExEUVNVWFFSWFVBT1hHR1lZIgRjaGFuKgVNRlJHRzIUx5tXg-Vym58og7aZw05IkoDvse8..c2lnbg")
	m := NewMessage(tk)
	m.TimelineContent("test", "", nil, []*MsgTimeItem{{Name: "123", Value: "123"}})
	var ctx pb.MsgContent
	if err := proto.Unmarshal(m.Content, &ctx); err != nil {
		t.Fatal("Unmarshal timeline content failed")
	}
	if len(ctx.TimeContent.TimeItems) != 0 {
		t.Fatal("Check time content failed")
	}
}

func TestMessageChannel(t *testing.T) {
	tk := &Token{}
	m := NewMessage(tk)
	if len(m.Channel) > 0 {
		t.Fatal("Check channel failed")
	}
	m.fixChannel()
	if !bytes.Equal(m.Channel, defaultChannel) {
		t.Fatal("Check default channel failed")
	}
	m = NewMessage(tk)
	m.TimelineContent("test", "", nil, []*MsgTimeItem{{Name: "123", Value: "123"}})
	if len(m.Channel) > 0 {
		t.Fatal("Check timeline channel failed")
	}
	m.fixChannel()
	if !bytes.Equal(m.Channel, timelineChannel) {
		t.Fatal("Check default timeline channel failed")
	}
}

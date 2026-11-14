package model

import (
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestAttachmentCaptionPreservesFile(t *testing.T) {
	msg := NewMessageDestination(nil, nil, nil).FileContent("files/synthetic", "report.txt", "old summary", 123, nil)
	msg.AttachmentCaption("Report", "new summary")
	var content pb.MsgContent
	if err := proto.Unmarshal(msg.Content, &content); err != nil {
		t.Fatal(err)
	}
	if content.Title != "Report" || content.Text != "new summary" || content.File != "files/synthetic" || content.Size != 123 || content.Filename != "report.txt" {
		t.Fatal("caption altered attachment data")
	}
	msg.AttachmentCaption("", "")
	if err := proto.Unmarshal(msg.Content, &content); err != nil || content.Title != "Report" || content.Text != "new summary" {
		t.Fatal("omitted caption erased metadata")
	}
}

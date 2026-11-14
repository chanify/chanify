package mcpnotify

import (
	"errors"
	"regexp"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

var mcpChannelCode = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)

// ChannelData converts readable codes; it never selects a default destination.
func ChannelData(code string) ([]byte, error) {
	c := &pb.Channel{Type: pb.ChanType_User, Name: code}
	switch code {
	case "sys.none":
		c = &pb.Channel{Type: pb.ChanType_Sys, Code: pb.ChanCode_Uncategorized}
	case "sys.timesets":
		c = &pb.Channel{Type: pb.ChanType_Sys, Code: pb.ChanCode_TimeSets}
	case "sys.device":
		return nil, errors.New("invalid channel code")
	default:
		if !mcpChannelCode.MatchString(code) {
			return nil, errors.New("invalid channel code")
		}
	}
	return proto.Marshal(c)
}

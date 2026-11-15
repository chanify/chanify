package mcpnotify

import "github.com/chanify/chanify/model"

// ChannelData converts readable codes; it never selects a default destination.
func ChannelData(code string) ([]byte, error) { return model.ChannelData(code) }

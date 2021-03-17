package logic

import (
	"net/url"

	"github.com/skip2/go-qrcode"
)

type Info struct {
	NodeId    string   `json:"nodeid"`
	Name      string   `json:"name,omitempty"`
	Version   string   `json:"version"`
	PublicKey string   `json:"pubkey"`
	Endpoint  string   `json:"endpoint,omitempty"`
	Features  []string `json:"features,omitempty"`
}

func (l *Logic) GetInfo() *Info {
	return &Info{
		NodeId:    l.NodeID,
		Name:      l.Name,
		Version:   l.Version,
		PublicKey: l.secKey.EncodePublicKey(),
		Endpoint:  l.Endpoint,
		Features:  l.Features,
	}
}

func (l *Logic) GetQRCode() []byte {
	qrcode, _ := qrcode.Encode("chanify://node?endpoint="+url.QueryEscape(l.Endpoint), qrcode.Medium, 256)
	return qrcode
}

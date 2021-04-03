package logic

import (
	"encoding/json"
	"net/url"

	"github.com/chanify/chanify/crypto"
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

func (l *Logic) InitInfo() {
	info := &Info{
		NodeId:    l.NodeID,
		Name:      l.Name,
		Version:   l.Version,
		PublicKey: l.secKey.EncodePublicKey(),
		Endpoint:  l.Endpoint,
		Features:  l.Features,
	}
	l.infoData, _ = json.Marshal(info)
	sign, _ := l.secKey.Sign(l.infoData)
	l.infoSign = crypto.Base64Encode.EncodeToString(sign)
}

func (l *Logic) GetInfo() ([]byte, string) {
	return l.infoData, l.infoSign
}

func (l *Logic) GetQRCode() []byte {
	qrcode, _ := qrcode.Encode("chanify://node?endpoint="+url.QueryEscape(l.Endpoint), qrcode.Medium, 256)
	return qrcode
}

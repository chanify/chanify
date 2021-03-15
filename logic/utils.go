package logic

import "encoding/base64"

var (
	base64Encode = base64.RawURLEncoding
)

func Encode(data []byte) string {
	return base64Encode.EncodeToString(data)
}

func Decode(s string) []byte {
	data, err := base64Encode.DecodeString(s)
	if err != nil {
		return nil
	}
	return data
}

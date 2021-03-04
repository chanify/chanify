package core

import "crypto/sha512"

func (c *Core) GetUserKey(uid string) ([]byte, error) {
	uidData, err := base32Encode.DecodeString(uid)
	if err != nil {
		return nil, err
	}
	h := sha512.New()
	h.Write(c.info.secret) // nolint: errcheck
	h.Write(uidData)       // nolint: errcheck
	return h.Sum(nil), nil
}

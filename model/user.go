package model

type User struct {
	Uid       string
	PublicKey []byte
	SecretKey []byte
	Flags     uint
}

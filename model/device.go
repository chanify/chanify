package model

// Device of user
type Device struct {
	Token   []byte
	Sandbox bool
	Type    int
}

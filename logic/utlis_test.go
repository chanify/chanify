package logic

import (
	"bytes"
	"testing"
)

func TestEncode(t *testing.T) {
	data := []byte("hello world!")
	s := Encode(data)
	if len(s) <= 0 {
		t.Fatal("Enocde failed")
	}
	if !bytes.Equal(data, Decode(s)) {
		t.Fatal("Decode failed")
	}
	if Decode("****") != nil {
		t.Fatal("Check decode invalid string failed")
	}
}

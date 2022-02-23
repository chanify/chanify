//go:build !test
// +build !test

//go:generate protoc --proto_path=./pb --go_out=./pb ./pb/pb.proto

package main

import "github.com/chanify/chanify/cmd"

func main() {
	cmd.Execute()
}

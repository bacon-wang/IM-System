package main

import "github.com/bacon-wang/IM-System/internal"

func main() {
	s := internal.NewServer("127.0.0.1", 8888)
	s.Start()
}

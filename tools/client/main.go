package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Parse command line arguments
	var serverIP = flag.String("ip", "127.0.0.1", "Server IP address")
	var serverPort = flag.String("port", "8888", "Server port")
	var username = flag.String("name", "bacon", "Your username (optional)")
	flag.Parse()

	// Show usage if help is requested
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("IM Client Tool")
		fmt.Println("Usage: im-client [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	fmt.Printf("Connecting to IM server at %s:%s...\n", *serverIP, *serverPort)

	// Create and start client
	client := NewClient(*serverIP, *serverPort, *username)
	if client == nil {
		fmt.Println("Failed to create client")
		os.Exit(1)
	}

	// Start the client
	client.Start()
}

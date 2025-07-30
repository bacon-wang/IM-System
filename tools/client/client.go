package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Client struct {
	ServerIp   string
	ServerPort string
	Name       string
	Addr       string
	conn       net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	quit       chan bool
	wg         sync.WaitGroup
}

func NewClient(serverIp string, serverPort string, name string) *Client {
	c := Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		Name:       name,
		quit:       make(chan bool),
	}

	// connect server
	conn, err := net.Dial("tcp", c.ServerIp+":"+c.ServerPort)
	if err != nil {
		panic(err)
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)

	return &c
}

// Start starts the client
func (c *Client) Start() {
	fmt.Println("Connected to IM server!")
	fmt.Println("Commands:")
	fmt.Println("  who                    - List online users")
	fmt.Println("  name                   - Show your current username")
	fmt.Println("  rename <newname>       - Change your username")
	fmt.Println("  to <username> <msg>    - Send private message")
	fmt.Println("  quit                   - Exit the client")
	fmt.Println("  Any other text will be broadcast to all users")
	fmt.Println()

	// rename
	err := c.sendMessage("rename " + c.Name)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
		c.quit <- true
		return
	}

	// Start goroutine to read messages from server
	c.wg.Add(1)
	go c.readMessages()

	// Start goroutine to handle user input
	c.wg.Add(1)
	go c.handleUserInput()

	// Wait for quit signal
	<-c.quit
	c.Close()
	c.wg.Wait()
}

// readMessages reads messages from the server
func (c *Client) readMessages() {
	defer c.wg.Done()

	for {
		select {
		case <-c.quit:
			return
		default:
			// Set read timeout to avoid blocking indefinitely
			c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			message, err := c.reader.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout, continue reading
				}
				if err.Error() != "EOF" {
					fmt.Printf("Error reading from server: %v\n", err)
				}
				return
			}

			// Clear read timeout
			c.conn.SetReadDeadline(time.Time{})

			// Process the message and update client name if needed
			c.processServerMessage(strings.TrimSpace(message))
		}
	}
}

// processServerMessage processes messages from server and updates client state
func (c *Client) processServerMessage(message string) {
	// Check if this is a successful rename response
	if strings.HasPrefix(message, "You've changed name to \"") && strings.HasSuffix(message, "\"") {
		// Extract the new name from the message
		// Format: You've changed name to "newname"
		start := strings.Index(message, "\"") + 1
		end := strings.LastIndex(message, "\"")
		if start > 0 && end > start {
			newName := message[start:end]
			c.Name = newName
			fmt.Printf("✓ Name updated to: %s\n", newName)
		}
	}

	// Display the message
	fmt.Print(message + "\n> ")
}

// GetCurrentName returns the current client name
func (c *Client) GetCurrentName() string {
	return c.Name
}

// handleUserInput handles user input
func (c *Client) handleUserInput() {
	defer c.wg.Done()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle quit command
		if input == "quit" || input == "exit" {
			c.quit <- true
			return
		}

		// Handle local commands
		if input == "name" {
			fmt.Printf("Current name: %s\n", c.GetCurrentName())
			continue
		}

		// Check if this is a rename command for better user feedback
		if strings.HasPrefix(input, "rename ") && len(input) > 7 {
			newName := input[7:]
			fmt.Printf("Attempting to rename to: %s\n", newName)
		}

		// Send message to server
		err := c.sendMessage(input)
		if err != nil {
			fmt.Printf("Error sending message: %v\n", err)
			c.quit <- true
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}

	c.quit <- true
}

// sendMessage sends a message to the server
func (c *Client) sendMessage(message string) error {
	_, err := c.writer.WriteString(message + "\n")
	if err != nil {
		return err
	}
	return c.writer.Flush()
}

// Close closes the client connection
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	fmt.Println("\nDisconnected from server. Goodbye!")
}

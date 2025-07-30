package main

import (
	"net"
	"strings"

	"go.uber.org/zap"
)

type User struct {
	Name   string
	Addr   string
	DataCh chan string
	conn   net.Conn

	server *Server
}

// create a user
func NewUser(conn net.Conn, server *Server) *User {
	addr := conn.RemoteAddr().String()
	u := &User{
		Name:   addr,
		Addr:   addr,
		DataCh: make(chan string),
		conn:   conn,
		server: server,
	}

	go u.listenMsg() // start listen DataCh

	return u
}

// listen DataCh and send data to user
func (u *User) listenMsg() {
	for {
		msg := <-u.DataCh
		u.conn.Write([]byte(msg + "\n"))
	}
}

// user online
func (u *User) Online() {
	// add user to online users
	u.server.OnlineUsersLock.Lock()
	u.server.OnlineUsers[u.Name] = u
	u.server.OnlineUsersLock.Unlock()

	// broadcast user online message
	u.server.broadCast(u, "login")
	u.server.Logger.Debug("User online",
		zap.String("name", u.Name),
		zap.String("addr", u.Addr))
}

// user offline
func (u *User) Offline() {
	u.server.OnlineUsersLock.Lock()
	delete(u.server.OnlineUsers, u.Name)
	u.server.OnlineUsersLock.Unlock()

	u.server.broadCast(u, "offline")
	u.server.Logger.Debug("User offline",
		zap.String("name", u.Name),
		zap.String("addr", u.Addr))

	u.conn.Close()
}

// send message to user
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

// MessageType represents different types of messages
type MessageType int

const (
	MessageTypeWho MessageType = iota
	MessageTypeRename
	MessageTypePrivate
	MessageTypeBroadcast
)

// MessageHandler defines the interface for message handlers
type MessageHandler interface {
	Handle(u *User, msg string) bool
}

// WhoHandler handles "who" command
type WhoHandler struct{}

func (h *WhoHandler) Handle(u *User, msg string) bool {
	if msg != "who" {
		return false
	}

	who := "online users: ["
	u.server.OnlineUsersLock.RLock()
	for username := range u.server.OnlineUsers {
		who += username + ", "
	}
	u.server.OnlineUsersLock.RUnlock()

	if len(who) > len("online users: [") {
		who = who[:len(who)-2]
	}
	who += "]\n"

	u.SendMsg(who)
	return true
}

// RenameHandler handles "rename" command
type RenameHandler struct{}

func (h *RenameHandler) Handle(u *User, msg string) bool {
	if len(msg) <= 7 || msg[:7] != "rename " {
		return false
	}

	newName := msg[7:]
	if u.server.isOnline(newName) {
		u.SendMsg("User name already exists\n")
		return true
	}

	u.ModifyName(newName)
	return true
}

// PrivateHandler handles private messages
type PrivateHandler struct{}

func (h *PrivateHandler) Handle(u *User, msg string) bool {
	if !strings.HasPrefix(msg, "to ") || len(msg) <= 3 {
		return false
	}

	split := strings.Split(msg, " ")
	if len(split) < 3 {
		u.SendMsg("Invalid message format, usage: to {username} {msg}\n")
		return true
	}

	username := split[1]
	if !u.server.isOnline(username) {
		u.SendMsg("User not online\n")
		return true
	}

	content := strings.Join(split[2:], " ")
	u.server.OnlineUsers[username].SendMsg(u.Name + " said to you: " + content + "\n")
	return true
}

// BroadcastHandler handles broadcast messages
type BroadcastHandler struct{}

func (h *BroadcastHandler) Handle(u *User, msg string) bool {
	u.server.broadCast(u, msg)
	return true
}

// MessageProcessor processes messages using registered handlers
type MessageProcessor struct {
	handlers []MessageHandler
}

// NewMessageProcessor creates a new message processor with default handlers
func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{
		handlers: []MessageHandler{
			&WhoHandler{},
			&RenameHandler{},
			&PrivateHandler{},
			&BroadcastHandler{}, // This should be last as it always returns true
		},
	}
}

// Process processes a message using the registered handlers
func (mp *MessageProcessor) Process(u *User, msg string) {
	for _, handler := range mp.handlers {
		if handler.Handle(u, msg) {
			return
		}
	}
}

// business logic
func (u *User) DoMessage(msg string) {
	u.server.Logger.Debug("recv from user",
		zap.String("name", u.Name),
		zap.String("message", msg))

	// Use message processor to handle the message
	processor := NewMessageProcessor()
	processor.Process(u, msg)
}

// modify username
func (u *User) ModifyName(newName string) {
	u.server.OnlineUsersLock.Lock()
	delete(u.server.OnlineUsers, u.Name)
	u.Name = newName
	u.server.OnlineUsers[u.Name] = u
	u.server.OnlineUsersLock.Unlock()

	u.SendMsg("You've changed name to \"" + newName + "\"\n")
}

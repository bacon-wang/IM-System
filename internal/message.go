package internal

import (
	"strings"
)

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

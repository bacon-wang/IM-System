package internal

import (
	"net"

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

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

// buissness logic
func (u *User) DoMessage(msg string) {
	u.server.Logger.Debug("recv from user",
		zap.String("name", u.Name),
		zap.String("message", msg))

	if msg == "who" { // who command
		who := "online users: ["

		u.server.OnlineUsersLock.RLock()
		for username := range u.server.OnlineUsers {
			who += username + ", "
		}
		u.server.OnlineUsersLock.RUnlock()

		who = who[:len(who)-2]
		who += "]\n"

		u.SendMsg(who)
	} else if len(msg) > 7 && msg[:7] == "rename " {
		// rename {new name}
		newName := msg[7:]
		if u.server.isOnline(newName) {
			u.SendMsg("User name already exists\n")
			return
		}

		u.ModifyName(newName)
	} else if strings.HasPrefix(msg, "to ") && len(msg) > 3 {
		// to {username} {msg}
		split := strings.Split(msg, " ")
		if len(split) < 3 {
			u.SendMsg("Invalid message format, useage: to {username} {msg}\n")
			return
		}

		username := split[1]
		if !u.server.isOnline(username) {
			u.SendMsg("User not online\n")
			return
		}
		content := split[2]
		u.server.OnlineUsers[username].SendMsg(u.Name + " said to you:" + content + "\n")
	} else { // normal data
		u.server.broadCast(u, msg)
	}
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

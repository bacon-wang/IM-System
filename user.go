package main

import "net"

type User struct {
	Name   string
	Addr   string
	DataCh chan string
	conn   net.Conn
}

// create a user
func NewUser(conn net.Conn) *User {
	addr := conn.RemoteAddr().String()
	u := &User{
		Name:   addr,
		Addr:   addr,
		DataCh: make(chan string),
		conn:   conn,
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

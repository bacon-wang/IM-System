package main

import (
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	Ip              string
	Port            int
	Logger          *zap.Logger
	OnlineUsers     map[string]*User
	OnlineUsersLock sync.RWMutex
	MsgCh           chan string // channel used to broadcast message to all clients
}

func NewServer(ip string, port int) *Server {
	logger, _ := zap.NewDevelopment()
	s := &Server{
		Ip:          ip,
		Port:        port,
		Logger:      logger,
		OnlineUsers: make(map[string]*User),
		MsgCh:       make(chan string),
	}
	return s
}

// receive message from global message channel and broadcast to all clients
func (s *Server) listenMsg() {
	for {
		msg, ok := <-s.MsgCh
		if !ok {
			s.Logger.Error("Failed to receive message from global message channel")
			return
		}

		s.OnlineUsersLock.Lock()
		for _, u := range s.OnlineUsers {
			u.DataCh <- msg
		}
		s.OnlineUsersLock.Unlock()
	}
}

// send message to global message channel
func (s *Server) broadCast(u *User, msg string) {
	fotmattedMsg := "[" + u.Addr + "]" + u.Name + ": " + msg
	s.MsgCh <- fotmattedMsg
}

func (s *Server) Handler(conn net.Conn) {
	u := NewUser(conn, s)

	u.Online()

	// listen that is user alive
	isAlive := make(chan bool)

	// process user message (with timeout kick)
	go func() {
		for {
			// read from user
			buf := make([]byte, 4096)

			n, err := u.conn.Read(buf)
			if n == 0 {
				u.server.Logger.Debug("read from user 0 bytes, close connection")
				u.Offline()
				return
			}

			if err != nil && err != io.EOF {
				u.server.Logger.Error("Failed to read from user", zap.Error(err))
				u.Offline()
				return
			}

			isAlive <- true

			msg := string(buf[:n-1]) // get rid off '\n'

			// do buissness logic
			u.DoMessage(msg)
		}
	}()

	for {
		select {
		case <-isAlive: // this case should be before timeout case, otherwise it will block
			// do nothing, just for activate select to reset timer
		case <-time.After(time.Second * 10):
			u.SendMsg("You've been kicked for timeout\n")
			u.Offline()
			return
		}
	}
}

func (s *Server) Start() {
	// listen
	listener, err := net.Listen("tcp", s.Ip+":"+strconv.Itoa(s.Port))
	if err != nil {
		s.Logger.Fatal("Failed to start server", zap.Error(err))
	}
	s.Logger.Info("Server started",
		zap.String("ip", s.Ip),
		zap.Int("port", s.Port))

	// close listen socket
	defer listener.Close()
	defer s.Logger.Sync()

	// start listen message channel
	go s.listenMsg()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			s.Logger.Error("Failed to accept connection", zap.Error(err))
			continue
		}
		s.Logger.Info("New connection accepted",
			zap.String("remote_addr", conn.RemoteAddr().String()))

		// do handler
		go s.Handler(conn)
	}
}

func (s *Server) isOnline(u *User) bool {
	s.OnlineUsersLock.RLock()
	_, ok := s.OnlineUsers[u.Name]
	s.OnlineUsersLock.RUnlock()
	return ok
}

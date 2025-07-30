package main

import (
	"io"
	"net"
	"strconv"
	"sync"

	"go.uber.org/zap"
)

type Server struct {
	Ip              string
	Port            int
	logger          *zap.Logger
	OnlineUsers     map[string]*User
	OnlineUsersLock sync.RWMutex
	MsgCh           chan string // channel used to broadcast message to all clients
}

func NewServer(ip string, port int) *Server {
	logger, _ := zap.NewDevelopment()
	s := &Server{
		Ip:          ip,
		Port:        port,
		logger:      logger,
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
			s.logger.Error("Failed to receive message from global message channel")
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
	// create user (listen message channel inside)
	u := NewUser(conn)

	// add user to online users
	s.OnlineUsersLock.Lock()
	s.OnlineUsers[u.Name] = u
	s.OnlineUsersLock.Unlock()

	// broadcast user online message
	s.broadCast(u, "login")
	s.logger.Debug("User online",
		zap.String("name", u.Name),
		zap.String("addr", u.Addr))

	// receive from user and broadcast (with timeout close)
	go func() {
		for {
			buf := make([]byte, 4096)

			n, err := u.conn.Read(buf)
			if n == 0 {
				msg := u.Name + " offline"
				s.broadCast(u, msg)
				s.logger.Debug(msg,
					zap.String("name", u.Name),
					zap.String("addr", u.Addr))
				delete(s.OnlineUsers, u.Name)
				return
			}

			if err != nil && err != io.EOF {
				s.logger.Error("Failed to read from user", zap.Error(err))
				u.conn.Close()
				delete(s.OnlineUsers, u.Name)
				return
			}

			s.broadCast(u, string(buf[:n-1])) // get rid off '\n'
		}
	}()

	select {}
}

func (s *Server) Start() {
	// listen
	listener, err := net.Listen("tcp", s.Ip+":"+strconv.Itoa(s.Port))
	if err != nil {
		s.logger.Fatal("Failed to start server", zap.Error(err))
	}
	s.logger.Info("Server started",
		zap.String("ip", s.Ip),
		zap.Int("port", s.Port))

	// close listen socket
	defer listener.Close()
	defer s.logger.Sync()

	// start listen message channel
	go s.listenMsg()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Error("Failed to accept connection", zap.Error(err))
			continue
		}
		s.logger.Info("New connection accepted",
			zap.String("remote_addr", conn.RemoteAddr().String()))

		// do handler
		go s.Handler(conn)
	}
}

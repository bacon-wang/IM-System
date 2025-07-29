package main

import (
	"net"
	"strconv"

	"go.uber.org/zap"
)

type Server struct {
	Ip     string
	Port   int
	logger *zap.Logger
}

func NewServer(ip string, port int) *Server {
	logger, _ := zap.NewDevelopment()
	s := &Server{
		Ip:     ip,
		Port:   port,
		logger: logger,
	}
	return s
}

func (s *Server) Handler(conn net.Conn) {
	// 业务逻辑
	s.logger.Info("connection established")
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

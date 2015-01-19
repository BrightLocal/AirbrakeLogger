package server

import (
	"bufio"
	"log"
	"net"
)

type Server struct {
	messageQueue chan []byte
}

func New(address string, messageQueue chan []byte) *Server {
	s := &Server{
		messageQueue,
	}
	go s.runListener(address)
	return s
}

func (s *Server) runListener(address string) error {
	socket, err := net.Listen("tcp4", address)
	if err != nil {
		return err
	}
	go s.listener(socket)
	return nil
}

func (s *Server) listener(socket net.Listener) {
	for {
		connection, err := socket.Accept()
		if err != nil {
			log.Print(err)
		}
		go s.handler(connection)
	}
}

func (s *Server) handler(connection net.Conn) {
	defer connection.Close()
	for {
		msg, err := bufio.NewReader(connection).ReadBytes(0)
		if err != nil {
			return
		}
		// Silently discard messages if chan is full
		if len(s.messageQueue) < cap(s.messageQueue) {
			s.messageQueue <- msg[:len(msg)-1]
		}
	}
}

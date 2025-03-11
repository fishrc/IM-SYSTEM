package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// online users list
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// broadcast channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

// listen Message and broadcast to all users
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message

		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "] " + user.Name + ": " + msg + "\n"
	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	// handle connection
	// fmt.Println("new connection created")
	user := NewUser(conn)
	// add new online user to list
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	// broadcast new online user
	s.BroadCast(user, "is online")
}

func (s *Server) Start() {
	// listen socket
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("listen error:", err)
	}
	defer listener.Close()

	go s.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		// do handler
		go s.Handler(conn)
	}
}

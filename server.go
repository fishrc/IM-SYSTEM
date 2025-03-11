package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
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

	user := NewUser(conn, s)
	user.Online()

	// current user is (not) alive
	isAlive := make(chan bool)

	// receive client's msg
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			msg := string(buf[:n-1])

			user.DoMessage(msg)

			isAlive <- true
		}
	}()

	for {
		select {
		case <-isAlive:
			// do nothing to go to next case, then update timer
		case <-time.After(time.Second * 10):
			conn.Write([]byte("you are kicked because of timeout\n"))

			conn.Close()
			close(user.C)
			// return
			runtime.Goexit()
		}
	}
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

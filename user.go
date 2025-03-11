package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// create a user api
func NewUser(conn net.Conn, server *Server) *User {

	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

func (u *User) Online() {
	// add new online user to list
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// broadcast new online user
	u.server.BroadCast(u, "is online")
}

func (u *User) Offline() {
	// remove offline user to list
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// broadcast offline user
	u.server.BroadCast(u, "is offline")
}

func (u *User) DoMessage(msg string) {
	if msg == "who" {
		// list online users
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "] " + user.Name + ": " + "is online\n"
			u.C <- onlineMsg
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := msg[7:]
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.C <- "name " + newName + " already exists\n"
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.Name = newName
			u.server.OnlineMap[u.Name] = u
			u.server.mapLock.Unlock()
			u.C <- "update name: " + newName + "\n"
		}
	} else if len(msg) > 3 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.C <- "user " + remoteName + " not exists\n"
		} else {
			content := strings.Split(msg, "|")[2]
			if content == "" {
				u.C <- "please say something\n"
			} else {
				remoteUser.C <- "from " + u.Name + ": " + content + "\n"
			}
		}
	} else {
		u.server.BroadCast(u, msg)
	}
}

// listen current user channel and send msg to client
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg))
	}
}

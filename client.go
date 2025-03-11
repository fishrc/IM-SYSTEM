package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func (c *Client) menu() bool {
	fmt.Println("1. Broadcast message")
	fmt.Println("2. Send private message")
	fmt.Println("3. Rename")
	fmt.Println("0. Logout")

	fmt.Scanln(&c.flag)
	if c.flag >= 0 && c.flag <= 3 {
		return true
	} else {
		fmt.Println("Error: Invalid option")
		return false
	}
}

func (c *Client) UpdateName() bool {
	fmt.Printf("Please type new name: ")
	fmt.Scanln(&c.Name)

	msg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return false
	}
	return true
}

func (c *Client) PublicChat() {
	var msg string
	fmt.Printf("Enter message to broadcast (type \"exit\" to exit): ")
	fmt.Scanln(&msg)

	for msg != "exit" {
		if len(msg) != 0 {
			_, err := c.conn.Write([]byte(msg + "\n"))
			if err != nil {
				fmt.Println("Error sending message:", err)
				break
			}
		}
		msg = ""
		fmt.Scanln(&msg)
	}
}

func (c *Client) SelectUser() {
	_, err := c.conn.Write([]byte("who\n"))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
}

func (c *Client) PrivateChat() {
	var remoteName string
	var msg string

	c.SelectUser()

	fmt.Printf("type user to send message to (type \"exit\" to exit): ")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Printf("Enter message to send to %s (type \"exit\" to exit): ", remoteName)
		fmt.Scanln(&msg)
		for msg != "exit" {
			if len(msg) != 0 {
				_, err := c.conn.Write([]byte("to|" + remoteName + "|" + msg + "\n"))
				if err != nil {
					fmt.Println("Error sending message:", err)
					break
				}
			}
			msg = ""
			fmt.Scanln(&msg)
		}
		remoteName = ""
		fmt.Printf("type user to send message to (type \"exit\" to exit): ")
		fmt.Scanln(&remoteName)
	}
}

func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {
			continue
		}
		switch c.flag {
		case 1:
			c.PublicChat()
			break
		case 2:
			c.PrivateChat()
			break
		case 3:
			c.UpdateName()
			break
		}
	}
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return nil
	}

	client.conn = conn

	return client
}

func (c *Client) DealResponse() {
	io.Copy(os.Stdout, c.conn)
}

var ServerIp string
var ServerPort int

func init() {
	flag.StringVar(&ServerIp, "ip", "localhost", "Server IP address")
	flag.IntVar(&ServerPort, "port", 3000, "Server port number")
}

func main() {
	flag.Parse()
	client := NewClient("localhost", 3000)
	if client == nil {
		fmt.Println("Failed to connect to server")
		return
	}
	fmt.Println("Connected to server")
	go client.DealResponse()
	client.Run()
}

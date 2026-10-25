package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"strings"
)

type ChatMessage struct {
	User string
	Body string
}

type ConnectArgs struct {
	Name string
}

func main() {
	c, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		fmt.Println("Connection failed!")
		return
	}

	in := bufio.NewReader(os.Stdin)
	fmt.Print("Your name: ")
	myName, _ := in.ReadString('\n')
	myName = strings.TrimSpace(myName)

	args := ConnectArgs{Name: myName}
	var answer string
	c.Call("ChatServer.Join", args, &answer)
	if answer == "EXISTS" {
		fmt.Println("Name already taken.")
		return
	}

	go func() {
		for {
			var msg ChatMessage
			err := c.Call("ChatServer.Receive", args, &msg)
			if err != nil {
				break
			}
			if msg.User == "System" {
				fmt.Printf("\n[NOTIFICATION]: %s\n> ", msg.Body)
			} else if msg.User != myName {
				fmt.Printf("\n[%s]: %s\n> ", msg.User, msg.Body)
			}
		}
	}()

	for {
		fmt.Print("> ")
		line, _ := in.ReadString('\n')
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}
		if line == "exit" {
			var ok bool
			c.Call("ChatServer.Leave", args, &ok)
			fmt.Println("Exiting chat...")
			break
		}

		m := ChatMessage{User: myName, Body: line}
		var success bool
		c.Call("ChatServer.Send", m, &success)
	}
}

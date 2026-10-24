package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"go-rpc-chat/chat"
)

type ChatClient struct {
	client   *rpc.Client
	username string
	lastID   int64
	mu       sync.Mutex
	done     chan struct{}
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Server Address (default localhost:12345): ")
	addr, _ := reader.ReadString('\n')
	addr = strings.TrimSpace(addr)
	if addr == "" {
		addr = "localhost:12345"
	}

	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	var username string
	for {
		fmt.Print("Enter unique username: ")
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)

		if username == "" {
			continue
		}

		args := chat.JoinArgs{Username: username}
		var reply chat.JoinReply
		err := client.Call("ChatServer.Join", &args, &reply)
		if err != nil {
			log.Fatalf("RPC Join Error: %v", err)
		}

		if reply.Success {
			fmt.Printf("Connected successfully as '%s'!\n\n", username)
			break
		} else {
			fmt.Printf("Error: %s. Please try again.\n", reply.Message)
		}
	}

	c := &ChatClient{
		client:   client,
		username: username,
		done:     make(chan struct{}),
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		c.Leave()
		fmt.Println("\nDisconnected cleanly.")
		os.Exit(0)
	}()

	go c.pollMessages()

	fmt.Println("--- CHAT ROOM INSTRUCTIONS ---")
	fmt.Println("Type your message and press ENTER.")
	fmt.Println("Commands: '/users' (list online users), '/quit' (exit)")
	fmt.Println("--------------------------------")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		if text == "/quit" {
			c.Leave()
			fmt.Println("Goodbye!")
			return
		} else if text == "/users" {
			c.listUsers()
		} else {
			c.sendMessage(text)
		}
	}
}

func (c *ChatClient) pollMessages() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			lastID := c.lastID
			c.mu.Unlock()

			args := chat.GetMessagesArgs{
				Username: c.username,
				LastID:   lastID,
			}
			var reply chat.GetMessagesReply

			err := c.client.Call("ChatServer.GetMessages", &args, &reply)
			if err != nil {
				return
			}

			if len(reply.Messages) > 0 {
				c.mu.Lock()
				for _, msg := range reply.Messages {
					if msg.ID > c.lastID {
						c.lastID = msg.ID
					}
					c.displayMessage(msg)
				}
				c.mu.Unlock()
			}
		}
	}
}

func (c *ChatClient) displayMessage(msg chat.Message) {
	timeStr := msg.Timestamp.Format("15:04:05")
	if msg.Type == chat.MsgSystem {
		fmt.Printf("\n[%s] %s\n> ", timeStr, msg.Content)
	} else {
		fmt.Printf("\n[%s] %s: %s\n> ", timeStr, msg.Sender, msg.Content)
	}
}

func (c *ChatClient) sendMessage(content string) {
	args := chat.SendMessageArgs{
		Sender:  c.username,
		Content: content,
	}
	var reply chat.SendMessageReply
	err := c.client.Call("ChatServer.SendMessage", &args, &reply)
	if err != nil {
		fmt.Printf("\n[Error sending message: %v]\n> ", err)
	}
}

func (c *ChatClient) listUsers() {
	var reply chat.ListUsersReply
	err := c.client.Call("ChatServer.ListUsers", &chat.ListUsersArgs{}, &reply)
	if err != nil {
		fmt.Printf("\n[Error fetching user list: %v]\n> ", err)
		return
	}
	fmt.Printf("\nOnline Users (%d): %s\n> ", len(reply.Users), strings.Join(reply.Users, ", "))
}

func (c *ChatClient) Leave() {
	close(c.done)
	args := chat.LeaveArgs{Username: c.username}
	var reply chat.LeaveReply
	_ = c.client.Call("ChatServer.Leave", &args, &reply)
}

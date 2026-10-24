package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go-rpc-chat/chat"
)

type ClientSession struct {
	Username string
	JoinedAt time.Time
	LastSeen time.Time
}

type ChatServer struct {
	mu         sync.Mutex
	clients    map[string]*ClientSession
	messages   []chat.Message
	msgCounter int64
}

func NewChatServer() *ChatServer {
	cs := &ChatServer{
		clients:  make(map[string]*ClientSession),
		messages: make([]chat.Message, 0),
	}
	go cs.reapInactiveClients()
	return cs
}

func (s *ChatServer) Join(args *chat.JoinArgs, reply *chat.JoinReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if args.Username == "" {
		reply.Success = false
		reply.Message = "Username cannot be empty"
		return nil
	}

	if _, exists := s.clients[args.Username]; exists {
		reply.Success = false
		reply.Message = "Username already taken"
		return nil
	}

	s.clients[args.Username] = &ClientSession{
		Username: args.Username,
		JoinedAt: time.Now(),
		LastSeen: time.Now(),
	}

	reply.Success = true
	reply.Message = "Successfully connected"

	s.addMessageLocked("Server", fmt.Sprintf("*** %s joined the chat ***", args.Username), chat.MsgSystem)
	log.Printf("[JOIN] User '%s' connected.", args.Username)
	return nil
}

func (s *ChatServer) Leave(args *chat.LeaveArgs, reply *chat.LeaveReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[args.Username]; exists {
		delete(s.clients, args.Username)
		s.addMessageLocked("Server", fmt.Sprintf("*** %s left the chat ***", args.Username), chat.MsgSystem)
		log.Printf("[LEAVE] User '%s' disconnected.", args.Username)
	}

	reply.Success = true
	return nil
}

func (s *ChatServer) SendMessage(args *chat.SendMessageArgs, reply *chat.SendMessageReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[args.Sender]; !exists {
		return errors.New("unauthorized: user not registered")
	}

	s.clients[args.Sender].LastSeen = time.Now()
	s.addMessageLocked(args.Sender, args.Content, chat.MsgUser)
	reply.Success = true
	return nil
}

func (s *ChatServer) GetMessages(args *chat.GetMessagesArgs, reply *chat.GetMessagesReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, exists := s.clients[args.Username]; exists {
		client.LastSeen = time.Now()
	}

	var newMsgs []chat.Message
	for _, msg := range s.messages {
		if msg.ID > args.LastID {
			if msg.Type == chat.MsgUser && msg.Sender == args.Username {
				continue
			}
			newMsgs = append(newMsgs, msg)
		}
	}

	reply.Messages = newMsgs
	return nil
}

func (s *ChatServer) ListUsers(args *chat.ListUsersArgs, reply *chat.ListUsersReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users := make([]string, 0, len(s.clients))
	for user := range s.clients {
		users = append(users, user)
	}
	reply.Users = users
	return nil
}

func (s *ChatServer) addMessageLocked(sender, content, msgType string) {
	s.msgCounter++
	msg := chat.Message{
		ID:        s.msgCounter,
		Sender:    sender,
		Content:   content,
		Type:      msgType,
		Timestamp: time.Now(),
	}
	s.messages = append(s.messages, msg)
}

func (s *ChatServer) reapInactiveClients() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for name, client := range s.clients {
			if now.Sub(client.LastSeen) > 15*time.Second {
				delete(s.clients, name)
				s.addMessageLocked("Server", fmt.Sprintf("*** %s disconnected (timeout) ***", name), chat.MsgSystem)
				log.Printf("[TIMEOUT] User '%s' timed out.", name)
			}
		}
		s.mu.Unlock()
	}
}

func main() {
	chatServer := NewChatServer()
	err := rpc.Register(chatServer)
	if err != nil {
		log.Fatalf("Failed to register RPC server: %v", err)
	}

	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatalf("Server listen error: %v", err)
	}
	defer listener.Close()

	log.Println("RPC Chat Server running on port :12345...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\nShutting down server gracefully...")
		listener.Close()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		go rpc.ServeConn(conn)
	}
}

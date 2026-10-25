package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

type ChatMessage struct {
	User string
	Body string
}

type ConnectArgs struct {
	Name string
}

type ChatServer struct {
	mu      sync.Mutex
	clients map[string]chan ChatMessage
}

func (s *ChatServer) Join(args ConnectArgs, res *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[args.Name]; exists {
		*res = "EXISTS"
		return nil
	}

	s.clients[args.Name] = make(chan ChatMessage, 100)
	*res = "WELCOME"

	s.broadcast(ChatMessage{User: "System", Body: args.Name + " has joined the chat."})
	fmt.Println("User joined:", args.Name)
	return nil
}

func (s *ChatServer) Leave(args ConnectArgs, ok *bool) error {
	s.mu.Lock()
	if ch, exists := s.clients[args.Name]; exists {
		close(ch)
		delete(s.clients, args.Name)
	}
	s.mu.Unlock()

	s.broadcast(ChatMessage{User: "System", Body: args.Name + " has left the chat."})
	fmt.Println("User left:", args.Name)
	*ok = true
	return nil
}

func (s *ChatServer) Send(m ChatMessage, ok *bool) error {
	s.broadcast(m)
	*ok = true
	return nil
}

func (s *ChatServer) Receive(args ConnectArgs, reply *ChatMessage) error {
	s.mu.Lock()
	ch, exists := s.clients[args.Name]
	s.mu.Unlock()

	if !exists {
		return fmt.Errorf("client not found")
	}

	msg := <-ch
	*reply = msg
	return nil
}

func (s *ChatServer) broadcast(m ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.clients {
		select {
		case ch <- m:
		default:
		}
	}
}

func main() {
	srv := &ChatServer{
		clients: make(map[string]chan ChatMessage),
	}
	rpc.Register(srv)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Server error")
		return
	}

	fmt.Println(" Server running ")
	http.Serve(l, nil)
}

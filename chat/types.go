package chat

import "time"

const (
	MsgUser   = "CHAT"
	MsgSystem = "SYSTEM"
)

type Message struct {
	ID        int64     `json:"id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

type JoinArgs struct {
	Username string
}

type JoinReply struct {
	Success bool
	Message string
}

type LeaveArgs struct {
	Username string
}

type LeaveReply struct {
	Success bool
}

type SendMessageArgs struct {
	Sender  string
	Content string
}

type SendMessageReply struct {
	Success bool
}

type GetMessagesArgs struct {
	Username string
	LastID   int64
}

type GetMessagesReply struct {
	Messages []Message
}

type ListUsersArgs struct{}

type ListUsersReply struct {
	Users []string
}

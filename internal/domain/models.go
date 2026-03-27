package domain

import "time"

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type ChatMessage struct {
	ID        string
	Content   string
	Role      MessageRole
	CreatedAt time.Time
}

type ChatRequest struct {
	Messages []ChatMessage
}

type ChatResponse struct {
	Response string
}

type SessionItem struct {
	ID        string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Messages  []ChatMessage
}

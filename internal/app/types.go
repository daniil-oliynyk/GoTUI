package app

import (
	"gotui/internal/chat"
	"gotui/internal/config"
	"gotui/internal/domain"
	"gotui/internal/store"
)

type ChatClientConfig = config.ChatClientConfig

type ChatMessage = domain.ChatMessage
type ChatRequest = domain.ChatRequest
type ChatResponse = domain.ChatResponse
type MessageRole = domain.MessageRole
type SessionItem = domain.SessionItem

const (
	MessageRoleUser      = domain.MessageRoleUser
	MessageRoleAssistant = domain.MessageRoleAssistant
)

type ChatClient = chat.Client
type ChatHistory = store.ChatHistory

func newChatClient(cfg ChatClientConfig) ChatClient {
	return chat.NewClient(cfg)
}

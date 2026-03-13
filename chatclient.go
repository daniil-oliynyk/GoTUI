package main

import (
	"log"
	"time"
)

type ChatClient interface {
	SendMessage(ChatRequest) (ChatResponse, error)
}

type chatClientImpl struct {
	Model string
}

func (c chatClientImpl) SendMessage(request ChatRequest) (ChatResponse, error) {
	log.Println("chatClientImpl.SendMessage().enter")
	time.Sleep(2 * time.Second)
	return ChatResponse{}, nil
}

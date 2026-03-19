package main

import (
	"context"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/conversations"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type ChatClient interface {
	SendMessage(ChatRequest) (ChatResponse, error)
	GetModels() ([]string, error)
}

type chatClientImpl struct {
	ctx          context.Context
	config       ChatClientConfig
	client       openai.Client
	conversation *conversations.Conversation
}

func newChatClient(config ChatClientConfig) ChatClient {
	c := &chatClientImpl{
		ctx:    context.Background(),
		config: config,
		client: openai.NewClient(
			option.WithAPIKey(config.APIKey),
		),
	}
	c.createConversation()
	return c
}

func (c *chatClientImpl) createConversation() {
	var err error
	c.conversation, err = c.client.Conversations.New(c.ctx, conversations.ConversationNewParams{})
	if err != nil {
		log.Println("chatClientImpl.createConversation().error", err)
	}
	log.Println("Created conversation:", c.conversation.ID)
}

func (c *chatClientImpl) SendMessage(request ChatRequest) (ChatResponse, error) {
	log.Println("chatClientImpl.SendMessage().enter")

	response, err := c.client.Responses.New(c.ctx, responses.ResponseNewParams{
		Model: c.config.Model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(request.Messages[len(request.Messages)-1].Content),
		},
		Conversation: responses.ResponseNewParamsConversationUnion{
			OfConversationObject: &responses.ResponseConversationParam{
				ID: c.conversation.ID,
			},
		},
	})
	if err != nil {
		log.Println("chatClientImpl.SendMessage().error", err)
		return ChatResponse{}, err
	}

	return ChatResponse{Response: response.OutputText()}, nil
}

func (c *chatClientImpl) GetModels() ([]string, error) {
	page, err := c.client.Models.List(c.ctx)
	if err != nil {
		log.Println("chatClientImpl.GetModels().error", err)
		return nil, err
	}

	var models []string
	for _, model := range page.Data {
		log.Println("Model:", model.ID)
		models = append(models, model.ID)
	}

	return models, nil

}

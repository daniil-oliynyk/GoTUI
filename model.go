package main

import (
	"log"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ChatMessage struct {
	Content string
	Role    MessageRole
}
type chatResponseMsg struct {
	message ChatMessage
}
type chatErrorMsg struct {
	err error
}
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type AppConfig struct {
	APIKey string
	Model  string
}

type ChatClient interface {
	SendMessage(ChatRequest) (ChatResponse, error)
}

type ChatRequest struct {
	Messages []ChatMessage
	Model    string
}

type ChatResponse struct {
	Response string
}

type chatClientImpl struct {
	Model string
}

func (c chatClientImpl) SendMessage(request ChatRequest) (ChatResponse, error) {
	log.Println("chatClientImpl.SendMessage().enter")
	return ChatResponse{}, nil
}

type Model struct {
	viewport     viewport.Model
	textinput    textinput.Model
	messages     []ChatMessage
	input        string
	pending      bool
	err          error
	width        int
	height       int
	cursor       int
	client       ChatClient
	config       AppConfig
	chatrequest  ChatRequest
	chatresponse ChatResponse
}

func newModel(config AppConfig) Model {
	vp := viewport.New(
		viewport.WithWidth(80),
		viewport.WithHeight(20),
	)

	ti := textinput.New()
	ti.Placeholder = "Placeholder"
	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(20)

	return Model{
		viewport:  vp,
		textinput: ti,
		messages:  []ChatMessage{},
		config:    config,
		client:    chatClientImpl{Model: config.Model},
	}
}

var botStyle = lipgloss.NewStyle().
	Align(lipgloss.Left).
	Padding(0, 1).
	Background(lipgloss.Color("238")).
	Foreground(lipgloss.Color("255")).
	Padding(0, 1).
	Margin(0, 10, 0, 0)

var userStyle = lipgloss.NewStyle().
	Align(lipgloss.Right).
	Padding(0, 1).
	Background(lipgloss.Color("62")).
	Foreground(lipgloss.Color("230")).
	Padding(0, 1).
	Margin(0, 0, 0, 10)

func (m Model) renderMessages() string {
	log.Println("renderMessages().enter")
	defer log.Println("renderMessages().exit")
	var renderedResult []string
	for _, msg := range m.messages {
		var rendered string

		if msg.Role == MessageRoleUser {
			rendered = userStyle.
				Width(50).
				Render(msg.Content)
		} else {
			rendered = botStyle.
				Width(m.viewport.Width()).
				Render(msg.Content)
		}

		renderedResult = append(renderedResult, rendered)
	}

	content := strings.Join(func() []string {

		var result []string
		for _, m := range renderedResult {
			result = append(result, m)
		}

		return result
	}(), "\n")

	return content
}

func (m Model) sendMessages() (ChatResponse, error) {
	log.Println("sendMessages().enter")
	defer log.Println("sendMessages().exit")

	request := ChatRequest{
		Messages: m.messages,
		Model:    m.config.Model,
	}
	m.chatrequest = request
	response, err := m.client.SendMessage(request)
	m.chatresponse = response
	return response, err

}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:

		inputHeight := 3
		m.viewport.SetHeight(msg.Height - inputHeight)
		m.viewport.SetWidth(msg.Width)
		m.textinput.SetWidth(msg.Width)

	case tea.KeyPressMsg:

		switch msg.String() {

		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			log.Println("Update().msg.enter")
			if m.textinput.Value() == "" {
				return m, nil
			}
			msg := ChatMessage{
				Content: m.textinput.Value(),
				Role:    MessageRoleUser,
			}
			log.Println("Update().msg.enter - added user message: " + msg.Content)
			m.messages = append(m.messages, msg)

			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			m.textinput.SetValue("")

			response, err := m.sendMessages()
			if err != nil {
				errorMessage := chatErrorMsg{err: err}
				log.Println("Update().msg.enter - error sending message: " + err.Error())
				return m, func() tea.Msg { return errorMessage }
			}
			chatMessage := chatResponseMsg{
				message: ChatMessage{
					Content: response.Response,
					Role:    MessageRoleAssistant,
				},
			}
			log.Println("Update().msg.enter - added assistant message: " + chatMessage.message.Content)
			return m, func() tea.Msg { return chatMessage }
		}

	case chatErrorMsg:
		log.Println("Update().msg.chatErrorMsg: " + msg.err.Error())
		return m, nil

	case chatResponseMsg:
		log.Println("Update().msg.chatResponseMsg: " + msg.message.Content)
		m.messages = append(m.messages, msg.message)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	}
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {

	var c *tea.Cursor
	if !m.textinput.VirtualCursor() {
		c = m.textinput.Cursor()
		c.Y += lipgloss.Height(m.viewport.View() + "\n")
	}

	str := lipgloss.JoinVertical(lipgloss.Top, "Header", m.viewport.View(), m.textinput.View(), "Footer")
	// if m.quitting {
	// 	str += "\n"
	// }

	v := tea.NewView(str)
	v.Cursor = c
	return v
}

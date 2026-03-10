package main

import (
	tea "charm.land/bubbletea/v2"
	// "charm.land/bubbles/v2"
	// "charm.land/lipgloss/v2"
)

type ChatMessage struct {
	Content string
	Role    MessageRole
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
	SendMessage(message string) (string, error)
}

type Model struct {
	messages []ChatMessage
	input    string
	pending  bool
	err      error
	width    int
	height   int
	cursor   int
	client   ChatClient
	config   AppConfig
}

type chatResponseMsg struct {
}
type chatErorMsg struct {
}

func newModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyPressMsg:

		switch msg.String() {

		case "ctrl+c", "esc":
			return m, tea.Quit
		}

	}

	return m, nil
}

func (m Model) View() tea.View {
	return tea.NewView("")
}

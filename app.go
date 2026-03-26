package main

import tea "charm.land/bubbletea/v2"

type screen int

const (
	screenSetup screen = iota
	screenChat
)

type navigateToChatMsg struct {
	config ChatClientConfig
}

type AppModel struct {
	screen      screen
	width       int
	height      int
	setupModel  SetupModel
	chatModel   ChatModel
	chatHistory ChatHistory
}

func newAppModel(defaultConfig ChatClientConfig, chatHistory ChatHistory) AppModel {
	return AppModel{
		screen:      screenSetup,
		setupModel:  newSetupModel(defaultConfig),
		chatHistory: chatHistory,
	}
}

func (a AppModel) Init() tea.Cmd {
	return a.setupModel.Init()
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		if a.screen == screenSetup {
			model, cmd := a.setupModel.Update(msg)
			a.setupModel = model.(SetupModel)
			return a, cmd
		}
		model, cmd := a.chatModel.Update(msg)
		a.chatModel = model.(ChatModel)
		return a, cmd

	case navigateToChatMsg:
		a.chatModel = newChatModel(msg.config, a.chatHistory)
		if a.width > 0 && a.height > 0 {
			model, _ := a.chatModel.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
			a.chatModel = model.(ChatModel)
		}
		a.screen = screenChat
		return a, nil
	}

	switch a.screen {
	case screenSetup:
		model, cmd := a.setupModel.Update(msg)
		a.setupModel = model.(SetupModel)
		return a, cmd

	case screenChat:
		model, cmd := a.chatModel.Update(msg)
		a.chatModel = model.(ChatModel)
		return a, cmd
	}

	return a, nil
}

func (a AppModel) View() tea.View {
	switch a.screen {
	case screenSetup:
		return a.setupModel.View()
	case screenChat:
		return a.chatModel.View()
	}
	return tea.NewView("")
}

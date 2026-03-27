package app

import (
	"errors"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type apiKeyLoadStep int

const (
	apiKeyLoadStepChoose apiKeyLoadStep = iota
	apiKeyLoadStepManual
)

type SetupModel struct {
	apiTokenInput textarea.Model
	err           error
	baseConfig    ChatClientConfig
	step          apiKeyLoadStep
	choiceIndex   int
	width         int
	height        int
}

func newSetupModel(defaultConfig ChatClientConfig) SetupModel {
	ti := textarea.New()
	ti.Placeholder = "sk-..."
	ti.Focus()
	ti.CharLimit = 200
	ti.SetWidth(40)
	ti.SetHeight(3)

	return SetupModel{
		apiTokenInput: ti,
		baseConfig:    defaultConfig,
		step:          apiKeyLoadStepChoose,
		choiceIndex:   0,
	}
}

func (m SetupModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		appInnerWidth := m.width - appStyle.GetHorizontalFrameSize()
		if appInnerWidth < 1 {
			appInnerWidth = 1
		}
		innerWidth := appInnerWidth - paneStyle.GetHorizontalFrameSize()
		if innerWidth < 1 {
			innerWidth = 1
		}

		maxInputWidth := innerWidth - 4
		if maxInputWidth < 1 {
			maxInputWidth = 1
		}
		if maxInputWidth > 72 {
			maxInputWidth = 72
		}
		m.apiTokenInput.SetWidth(maxInputWidth)
		m.apiTokenInput.SetHeight(3)
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.step == apiKeyLoadStepChoose {
				if m.choiceIndex > 0 {
					m.choiceIndex--
				}
				return m, nil
			}
		case "down", "j":
			if m.step == apiKeyLoadStepChoose {
				if m.choiceIndex < 1 {
					m.choiceIndex++
				}
				return m, nil
			}
		case "enter":
			if m.step == apiKeyLoadStepChoose {
				m.err = nil
				if m.choiceIndex == 0 {
					if strings.TrimSpace(m.baseConfig.APIKey) == "" {
						m.err = errors.New("No API key found in environment (.env). Set API_KEY and try again, or choose manual entry.")
						return m, nil
					}
					return m, func() tea.Msg {
						return navigateToChatMsg{config: m.baseConfig}
					}
				}

				m.step = apiKeyLoadStepManual
				m.apiTokenInput.Focus()
				return m, nil
			}

			apiKey := strings.TrimSpace(m.apiTokenInput.Value())
			if apiKey == "" {
				m.err = errors.New("API token is required")
				return m, nil
			}
			m.err = nil
			config := m.baseConfig
			config.APIKey = apiKey
			return m, func() tea.Msg {
				return navigateToChatMsg{config: config}
			}
		}
	}

	if m.step == apiKeyLoadStepManual {
		var cmd tea.Cmd
		m.apiTokenInput, cmd = m.apiTokenInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SetupModel) View() tea.View {
	appInnerWidth := m.width - appStyle.GetHorizontalFrameSize()
	appInnerHeight := m.height - appStyle.GetVerticalFrameSize()
	if appInnerWidth < 1 {
		appInnerWidth = 1
	}
	if appInnerHeight < 1 {
		appInnerHeight = 1
	}

	sectionWidth := appInnerWidth

	header := headerStyle.Width(sectionWidth - headerStyle.GetHorizontalFrameSize()).Render("chatTUI")
	footer := footerStyle.Width(sectionWidth - footerStyle.GetHorizontalFrameSize()).Render("Enter to continue | Ctrl+C or esc to exit")

	innerWidth := sectionWidth - paneStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}

	var form string
	if m.step == apiKeyLoadStepChoose {
		prompt := labelStyle.Render("How would you like to load your OpenAI API key?")
		opt0 := "Load from .env (API_KEY)"
		opt1 := "Paste manually"

		selectedStyle := optStyle.Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
		line0 := optStyle.Render(opt0)
		line1 := optStyle.Render(opt1)
		if m.choiceIndex == 0 {
			line0 = selectedStyle.Render(opt0)
		}
		if m.choiceIndex == 1 {
			line1 = selectedStyle.Render(opt1)
		}

		hint := hintStyle.Render("Use ↑/↓ then Enter")
		form = formStyle.Render(lipgloss.JoinVertical(lipgloss.Left, prompt, "", line0, line1, "", hint))
	} else {
		label := labelStyle.Render("OpenAI API token")
		hint := hintStyle.Render("Your key is only used to create the chat session")
		form = formStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			label,
			m.apiTokenInput.View(),
			hint,
		))
	}

	errLine := ""
	if m.err != nil {
		errLine = errLineStyle.Render(m.err.Error())
	}

	if errLine != "" {
		form = formStyle.Render(lipgloss.JoinVertical(lipgloss.Left, form, "", errLine))
	}

	paneTotalHeight := appInnerHeight - lipgloss.Height(header) - lipgloss.Height(footer)
	if paneTotalHeight < 1 {
		paneTotalHeight = 1
	}
	paneInnerHeight := paneTotalHeight - paneStyle.GetVerticalFrameSize()
	if paneInnerHeight < 1 {
		paneInnerHeight = 1
	}
	form = lipgloss.Place(innerWidth, paneInnerHeight, lipgloss.Center, lipgloss.Center, form)
	centerPane := paneStyle.Width(innerWidth).Render(form)

	content := lipgloss.JoinVertical(lipgloss.Left, header, centerPane, footer)
	return tea.NewView(appStyle.Render(content))
}

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ChatMessage struct {
	ID        string
	Content   string
	Role      MessageRole
	CreatedAt time.Time
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

type ChatRequest struct {
	Messages []ChatMessage
}

type ChatResponse struct {
	Response string
}

type ChatModel struct {
	spinner              spinner.Model
	viewport             viewport.Model
	textinput            textinput.Model
	messages             []ChatMessage
	input                string
	pending              bool
	err                  error
	width                int
	height               int
	cursor               int
	client               ChatClient
	chatClientConfig     ChatClientConfig
	presetModels         []string
	allModels            []string
	modelList            []string
	allModelsExpanded    bool
	modelPickerOpen      bool
	modelPickerIndex     int
	modelPickerOffset    int
	sessions             []SessionItem
	selectedSession      int
	sessionListOffset    int
	sessionDeleteConfirm bool
	sessionDeleteTarget  int
	chatrequest          ChatRequest
	chatresponse         ChatResponse
}

type layoutSections struct {
	header   string
	status   string
	composer string
	footer   string
}

func newChatModel(config ChatClientConfig) ChatModel {
	presetModels := []string{
		"gpt-5-nano-2025-08-07",
		"gpt-5.4",
		"gpt-5.4-mini",
		"gpt-5.4-nano",
		"gpt-5.3-codex",
	}

	modelPickerIndex := selectedModelIndex(config.Model, presetModels)

	vp := viewport.New(
		viewport.WithWidth(80),
		viewport.WithHeight(20),
	)
	vp.MouseWheelEnabled = true

	s := spinner.New()
	s.Spinner = spinner.Points

	ti := textinput.New()
	ti.Placeholder = "Ask anything"
	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(20)

	m := ChatModel{
		spinner:           s,
		viewport:          vp,
		textinput:         ti,
		pending:           false,
		messages:          []ChatMessage{},
		chatClientConfig:  config,
		client:            newChatClient(config),
		presetModels:      presetModels,
		modelList:         presetModels,
		modelPickerIndex:  modelPickerIndex,
		allModelsExpanded: false,
		sessions: []SessionItem{
			{Title: "Session 1"},
			{Title: "Session 2"},
			{Title: "Session 3"},
			{Title: "Session 4"},
			{Title: "Session 5"},
			{Title: "Session 6"},
			{Title: "Session 7"},
			{Title: "Session 8"},
		},
		selectedSession: 0,
	}

	allModels, err := m.client.GetModels()
	if err != nil {
		log.Printf("Failed to get models: %v", err)
	} else {
		m.allModels = allModels
	}
	return m
}

func selectedModelIndex(currentModel string, availableModels []string) int {
	for i, model := range availableModels {
		if model == currentModel {
			return i
		}
	}

	return 0
}

func (m ChatModel) selectedSessionTitle() string {
	if len(m.sessions) == 0 {
		return "No Session"
	}

	if m.selectedSession < 0 || m.selectedSession >= len(m.sessions) {
		return "No Session"
	}

	return m.sessions[m.selectedSession].Title
}

func (m *ChatModel) ensureSessionSelectionInBounds() {
	if len(m.sessions) == 0 {
		m.selectedSession = 0
		m.sessionListOffset = 0
		return
	}

	if m.selectedSession < 0 {
		m.selectedSession = 0
	}
	if m.selectedSession > len(m.sessions)-1 {
		m.selectedSession = len(m.sessions) - 1
	}

	visible := m.sessionVisibleCount()
	if m.sessionListOffset > m.selectedSession {
		m.sessionListOffset = m.selectedSession
	}
	if m.selectedSession >= m.sessionListOffset+visible {
		m.sessionListOffset = m.selectedSession - visible + 1
	}

	maxOffset := len(m.sessions) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.sessionListOffset < 0 {
		m.sessionListOffset = 0
	}
	if m.sessionListOffset > maxOffset {
		m.sessionListOffset = maxOffset
	}
}

func (m *ChatModel) addSession() {
	next := len(m.sessions) + 1
	m.sessions = append(m.sessions, SessionItem{Title: fmt.Sprintf("Session %d", next)})
	m.selectedSession = len(m.sessions) - 1
	m.ensureSessionSelectionInBounds()
}

func (m *ChatModel) deleteSelectedSession() {
	if len(m.sessions) <= 1 {
		m.sessionDeleteConfirm = false
		return
	}

	target := m.sessionDeleteTarget
	if target < 0 || target >= len(m.sessions) {
		target = m.selectedSession
	}

	m.sessions = append(m.sessions[:target], m.sessions[target+1:]...)
	if m.selectedSession >= len(m.sessions) {
		m.selectedSession = len(m.sessions) - 1
	}
	m.sessionDeleteConfirm = false
	m.ensureSessionSelectionInBounds()
}

func (m ChatModel) sessionVisibleCount() int {
	return 6
}

func (m ChatModel) sessionsSidebarWidth(totalWidth int) int {
	sidebarWidth := 36
	if totalWidth < 70 {
		sidebarWidth = totalWidth / 2
	}
	if sidebarWidth < 24 {
		sidebarWidth = 24
	}
	if sidebarWidth > totalWidth-1 {
		sidebarWidth = totalWidth - 1
	}
	if sidebarWidth < 1 {
		sidebarWidth = 1
	}

	return sidebarWidth
}

func (m ChatModel) renderMessages() string {
	log.Println("renderMessages().enter")
	defer log.Println("renderMessages().exit")
	var renderedResult []string

	paneInnerWidth := m.transcriptPaneWidth()
	conversationWidth := m.conversationWidth(paneInnerWidth)
	conversationLaneWidth := m.conversationLaneWidth(conversationWidth)
	assistantBubbleWidth := m.assistantBubbleWidth(conversationLaneWidth)
	userBubbleWidth := m.userBubbleWidth(conversationLaneWidth)

	for _, msg := range m.messages {
		var rendered string

		if msg.Role == MessageRoleUser {
			bubble := userStyle.
				Width(userBubbleWidth - userStyle.GetHorizontalFrameSize()).
				Render(msg.Content)
			row := lipgloss.PlaceHorizontal(conversationLaneWidth, lipgloss.Right, bubble)
			rendered = lipgloss.PlaceHorizontal(paneInnerWidth, lipgloss.Center, row)
		} else {
			bubble := botStyle.
				Width(assistantBubbleWidth - botStyle.GetHorizontalFrameSize()).
				Render(msg.Content)
			row := lipgloss.PlaceHorizontal(conversationLaneWidth, lipgloss.Left, bubble)
			rendered = lipgloss.PlaceHorizontal(paneInnerWidth, lipgloss.Center, row)
		}

		renderedResult = append(renderedResult, rendered)
	}

	content := strings.Join(renderedResult, "\n\n")

	return content
}

func sendMessages(m ChatModel) tea.Cmd {
	log.Println("sendMessages().enter")
	defer log.Println("sendMessages().exit")

	return func() tea.Msg {

		request := ChatRequest{
			Messages: m.messages,
		}
		m.chatrequest = request
		response, err := m.client.SendMessage(request)
		m.chatresponse = response
		if err != nil {
			errorMessage := chatErrorMsg{err: err}
			log.Println("m.sendMessages() - error sending message: " + err.Error())
			return errorMessage
		}
		chatMessage := chatResponseMsg{
			message: ChatMessage{
				Content: response.Response,
				Role:    MessageRoleAssistant,
			},
		}
		log.Println("m.sendMessages() - received message: " + chatMessage.message.Content)
		return chatMessage
	}
}

func (m ChatModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.MouseMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:

		m.width = msg.Width
		m.height = msg.Height

		appInnerWidth := m.width - appStyle.GetHorizontalFrameSize()
		appInnerHeight := m.height - appStyle.GetVerticalFrameSize()

		if appInnerWidth < 1 {
			appInnerWidth = 1
		}
		if appInnerHeight < 1 {
			appInnerHeight = 1
		}

		sectionWidth := appInnerWidth
		paneInnerWidth := sectionWidth - paneStyle.GetHorizontalFrameSize()
		rightPaneInnerWidth := paneInnerWidth - m.sessionsSidebarWidth(paneInnerWidth) - 2
		composerInnerWidth := sectionWidth - m.currentComposerStyle().GetHorizontalFrameSize()

		if paneInnerWidth < 1 {
			paneInnerWidth = 1
		}
		if rightPaneInnerWidth < 1 {
			rightPaneInnerWidth = 1
		}
		if composerInnerWidth < 1 {
			composerInnerWidth = 1
		}

		m.viewport.SetWidth(rightPaneInnerWidth)
		m.textinput.SetWidth(composerInnerWidth)

		sections := m.renderLayoutSections(sectionWidth)
		headerHeight := lipgloss.Height(sections.header)
		statusHeight := lipgloss.Height(sections.status)
		footerHeight := lipgloss.Height(sections.footer)
		composerHeight := lipgloss.Height(sections.composer)

		paneTotalHeight := appInnerHeight - headerHeight - statusHeight - composerHeight - footerHeight
		if paneTotalHeight < 1 {
			paneTotalHeight = 1
		}

		paneInnerHeight := paneTotalHeight - paneStyle.GetVerticalFrameSize()
		if paneInnerHeight < 1 {
			paneInnerHeight = 1
		}

		m.viewport.SetHeight(paneInnerHeight)
		m.viewport.SetContent(m.renderMessages())

	case tea.KeyPressMsg:
		if m.modelPickerOpen {

			switch msg.String() {
			case "esc":
				if m.allModelsExpanded {
					m.modelList = m.presetModels
					m.allModelsExpanded = false
					m.modelPickerIndex = selectedModelIndex(m.chatClientConfig.Model, m.modelList)
					m.modelPickerOffset = 0
					return m, nil
				}
				m.modelPickerOpen = false
				return m, nil

			case "up", "k":
				if len(m.modelList) == 0 {
					return m, nil
				}
				if m.modelPickerIndex > 0 {
					m.modelPickerIndex--
				}
				if m.modelPickerIndex < m.modelPickerOffset {
					m.modelPickerOffset = m.modelPickerIndex
				}
				return m, nil

			case "down", "j":
				if len(m.modelList) == 0 {
					return m, nil
				}
				if m.modelPickerIndex < len(m.modelList)-1 {
					m.modelPickerIndex++
				}
				visibleCount := m.modelPickerVisibleCount(m.viewport.Height())
				if m.modelPickerIndex >= m.modelPickerOffset+visibleCount {
					m.modelPickerOffset = m.modelPickerIndex - visibleCount + 1
				}
				return m, nil

			case "pgup":
				if len(m.modelList) == 0 {
					return m, nil
				}
				visibleCount := m.modelPickerVisibleCount(m.viewport.Height())
				m.modelPickerIndex -= visibleCount
				if m.modelPickerIndex < 0 {
					m.modelPickerIndex = 0
				}
				if m.modelPickerIndex < m.modelPickerOffset {
					m.modelPickerOffset = m.modelPickerIndex
				}
				return m, nil

			case "pgdown":
				if len(m.modelList) == 0 {
					return m, nil
				}
				visibleCount := m.modelPickerVisibleCount(m.viewport.Height())
				m.modelPickerIndex += visibleCount
				if m.modelPickerIndex > len(m.modelList)-1 {
					m.modelPickerIndex = len(m.modelList) - 1
				}
				if m.modelPickerIndex >= m.modelPickerOffset+visibleCount {
					m.modelPickerOffset = m.modelPickerIndex - visibleCount + 1
				}
				return m, nil

			case "e":
				if len(m.allModels) > 0 {
					m.modelList = m.allModels
					m.allModelsExpanded = true
				}
				log.Println("Model list updated to all models: ", len(m.modelList))
				m.modelPickerIndex = selectedModelIndex(m.chatClientConfig.Model, m.modelList)
				m.modelPickerOffset = 0
				visibleCount := m.modelPickerVisibleCount(m.viewport.Height())
				if m.modelPickerIndex >= visibleCount {
					m.modelPickerOffset = m.modelPickerIndex - visibleCount + 1
				}
				return m, nil

			case "enter":
				if len(m.modelList) == 0 {
					return m, nil
				}
				m.chatClientConfig.Model = m.modelList[m.modelPickerIndex]
				m.client = newChatClient(m.chatClientConfig)
				m.modelPickerOpen = false
				return m, nil
			}

			return m, nil
		}

		if m.sessionDeleteConfirm {
			switch msg.String() {
			case "esc":
				m.sessionDeleteConfirm = false
				return m, nil
			case "alt+d", "enter":
				m.deleteSelectedSession()
				return m, nil
			default:
				return m, nil
			}
		}

		switch msg.String() {

		case "ctrl+c", "esc":
			return m, tea.Quit

		case "ctrl+o":
			m.modelList = m.presetModels
			m.modelPickerIndex = selectedModelIndex(m.chatClientConfig.Model, m.modelList)
			m.modelPickerOffset = 0
			visibleCount := m.modelPickerVisibleCount(m.viewport.Height())
			if m.modelPickerIndex >= visibleCount {
				m.modelPickerOffset = m.modelPickerIndex - visibleCount + 1
			}
			m.modelPickerOpen = true
			return m, nil

		case "ctrl+n":
			m.addSession()
			return m, nil

		case "alt+s":
			if len(m.sessions) == 0 {
				return m, nil
			}
			m.selectedSession++
			if m.selectedSession >= len(m.sessions) {
				m.selectedSession = 0
				m.sessionListOffset = 0
			}
			m.ensureSessionSelectionInBounds()
			return m, nil

		case "alt+w":
			if len(m.sessions) == 0 {
				return m, nil
			}
			m.selectedSession--
			if m.selectedSession < 0 {
				m.selectedSession = len(m.sessions) - 1
			}
			m.ensureSessionSelectionInBounds()
			return m, nil

		case "alt+d":
			if len(m.sessions) <= 1 {
				return m, nil
			}
			m.sessionDeleteConfirm = true
			m.sessionDeleteTarget = m.selectedSession
			return m, nil

		case "enter":
			if m.pending {
				return m, nil
			}

			log.Println("Update().msg.enter")
			if m.textinput.Value() == "" {
				return m, nil
			}
			m.pending = true

			msg := ChatMessage{
				Content: m.textinput.Value(),
				Role:    MessageRoleUser,
			}
			log.Println("Update().msg.enter - added user message: " + msg.Content)
			m.messages = append(m.messages, msg)

			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			m.textinput.SetValue("")

			return m, tea.Batch(
				m.spinner.Tick,
				sendMessages(m),
			)

		case "up":
			m.viewport.ScrollUp(5)
			return m, nil

		case "down":
			m.viewport.ScrollDown(5)
			return m, nil
		}

	case chatErrorMsg:
		log.Println("Update().msg.chatErrorMsg.Content: " + msg.err.Error())
		m.pending = false
		return m, nil

	case chatResponseMsg:
		log.Println("Update().msg.chatResponseMsg.Content: " + msg.message.Content)
		m.pending = false
		m.messages = append(m.messages, msg.message)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		log.Println("Update().msg.chatResponseMsg message added")
		return m, nil

	case spinner.TickMsg:
		if m.pending {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	}
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m ChatModel) View() tea.View {
	appInnerWidth := m.width - appStyle.GetHorizontalFrameSize()
	if appInnerWidth < 1 {
		appInnerWidth = 1
	}

	sections := m.renderLayoutSections(appInnerWidth)
	pane := m.renderPane(appInnerWidth)

	var c *tea.Cursor
	if !m.textinput.VirtualCursor() {
		c = m.textinput.Cursor()
		composerStyle := m.currentComposerStyle()
		composerInnerWidth := appInnerWidth - composerStyle.GetHorizontalFrameSize()
		if composerInnerWidth < 1 {
			composerInnerWidth = 1
		}

		aboveComposer := lipgloss.Height(
			lipgloss.JoinVertical(
				lipgloss.Left,
				sections.header,
				pane,
				sections.status,
			),
		)

		composerTopOffset := composerStyle.GetVerticalFrameSize() / 2
		composerLeftOffset := composerStyle.GetHorizontalFrameSize() / 2
		appTopOffset := appStyle.GetVerticalFrameSize() / 2
		appLeftOffset := appStyle.GetHorizontalFrameSize() / 2

		c.Y += appTopOffset + aboveComposer + composerTopOffset
		c.X += appLeftOffset + composerLeftOffset
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		sections.header,
		pane,
		sections.status,
		sections.composer,
		sections.footer,
	)

	str := appStyle.Render(content)
	v := tea.NewView(str)
	if m.modelPickerOpen {
		c = nil
	}
	v.Cursor = c
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m ChatModel) renderHeader(width int) string {
	innerWidth := width - headerStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	return headerStyle.Width(innerWidth).Render("Active Model: " + m.chatClientConfig.Model + " | Session: " + m.selectedSessionTitle())
}

func (m ChatModel) renderStatus(width int) string {
	statusText := ""
	if m.pending {
		statusText = "Thinking " + m.spinner.View()
	}
	if m.err != nil {
		statusText = "Error: " + m.err.Error()
	}
	innerWidth := width - statusStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	status := statusStyle.Width(innerWidth).Render(statusText)
	return status
}

func (m ChatModel) renderPane(width int) string {
	innerWidth := width - paneStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	content := m.viewport.View()
	if m.modelPickerOpen {
		content = m.renderModelPicker(innerWidth, m.viewport.Height())
	} else {
		sidebarWidth := m.sessionsSidebarWidth(innerWidth)
		mainWidth := innerWidth - sidebarWidth - 2
		if mainWidth < 1 {
			mainWidth = 1
		}

		sidebar := m.renderSessionsSidebar(sidebarWidth, m.viewport.Height())

		mainContent := m.viewport.View()
		if len(m.messages) == 0 {
			mainContent = m.renderSessionPlaceholder(mainWidth, m.viewport.Height())
		}

		right := lipgloss.NewStyle().Width(mainWidth).Render(mainContent)
		content = lipgloss.JoinHorizontal(lipgloss.Top, " ", sidebar, " ", right)
	}
	return paneStyle.Width(innerWidth).Render(content)
}

func (m ChatModel) renderComposer(width int) string {
	composerStyle := m.currentComposerStyle()
	innerWidth := width - composerStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,

		m.textinput.View(),
	)

	return composerStyle.Width(innerWidth).Render(body)
}

func (m ChatModel) renderFooter(width int) string {
	innerWidth := width - footerStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	if m.modelPickerOpen {
		return footerStyle.Width(innerWidth).Render("↑/↓ choose | Enter select | Esc close")
	}

	if m.sessionDeleteConfirm {
		return footerStyle.Width(innerWidth).Render("Delete session? Enter/alt+d confirm | Esc cancel")
	}

	return footerStyle.Width(innerWidth).Render("Enter send | Ctrl+N new session | alt+d delete | alt+w session up | alt+s session down | Ctrl+O models | Ctrl+C or esc exit")
}

func (m ChatModel) renderSessionsSidebar(width, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	rows := []string{
		sessionSidebarTitleStyle.Render("Sessions"),
		sessionSidebarActionStyle.Render("+ New Session (Ctrl+N)"),
		"",
	}

	if len(m.sessions) == 0 {
		rows = append(rows, hintStyle.Render("No sessions"))
	} else {
		visible := m.sessionVisibleCount()
		start := m.sessionListOffset
		if start < 0 {
			start = 0
		}
		maxStart := len(m.sessions) - visible
		if maxStart < 0 {
			maxStart = 0
		}
		if start > maxStart {
			start = maxStart
		}
		end := start + visible
		if end > len(m.sessions) {
			end = len(m.sessions)
		}

		for i := start; i < end; i++ {
			line := optStyle.Render(m.sessions[i].Title)
			if i == m.selectedSession {
				line = sessionSelectedStyle.Render(m.sessions[i].Title)
			}
			rows = append(rows, line)
		}

		rows = append(rows, "", hintStyle.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(m.sessions))))
	}

	if m.sessionDeleteConfirm {
		rows = append(rows, "", sessionDeletePromptStyle.Render("Delete selected session?"), hintStyle.Render("Enter/alt+d confirm | Esc cancel"))
	}

	innerWidth := width - sessionsSidebarStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := height - sessionsSidebarStyle.GetVerticalFrameSize()
	if innerHeight < 1 {
		innerHeight = 1
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return sessionsSidebarStyle.Width(innerWidth).Height(innerHeight).Render(content)
}

func (m ChatModel) renderSessionPlaceholder(width, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	title := labelStyle.Width(width).Align(lipgloss.Center).Render(m.selectedSessionTitle())
	subtitle := emptyStateSubtitleStyle.Width(width).Align(lipgloss.Center).Render("No messages yet")
	body := lipgloss.JoinVertical(lipgloss.Center, title, "", subtitle)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, body)
}

func (m ChatModel) renderModelPicker(width, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	selectedStyle := optStyle.
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230"))

	rows := make([]string, 0, len(m.modelList)+6)
	rows = append(rows, labelStyle.Render("Choose OpenAI model"), "")

	if len(m.modelList) == 0 {
		rows = append(rows, hintStyle.Render("No models available"))
	} else {
		visibleCount := m.modelPickerVisibleCount(height)
		start := m.modelPickerOffset
		if start < 0 {
			start = 0
		}
		maxStart := len(m.modelList) - visibleCount
		if maxStart < 0 {
			maxStart = 0
		}
		if start > maxStart {
			start = maxStart
		}

		end := start + visibleCount
		if end > len(m.modelList) {
			end = len(m.modelList)
		}

		for i := start; i < end; i++ {
			model := m.modelList[i]
			line := optStyle.Render(model)
			if i == m.modelPickerIndex {
				line = selectedStyle.Render(model)
			}
			rows = append(rows, line)
		}

		rangeHint := fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(m.modelList))
		rows = append(rows, "", hintStyle.Render(rangeHint))
	}

	rows = append(rows, "", hintStyle.Render("Use ↑/↓ or PgUp/PgDn then Enter"))
	if !m.allModelsExpanded {
		rows = append(rows, hintStyle.Render("Press 'e' to expand model list"))
	}
	rows = append(rows, hintStyle.Render("Only chat is supported, no audio/video/image generation"))

	popup := formStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, popup)
}

func (m ChatModel) modelPickerVisibleCount(height int) int {
	visibleCount := height - formStyle.GetVerticalFrameSize() - 7
	if visibleCount < 1 {
		visibleCount = 1
	}

	return visibleCount
}

func (m ChatModel) currentComposerStyle() lipgloss.Style {
	if m.textinput.Focused() {
		return composerFocusedStyle
	}

	return composerBlurredStyle
}

func (m ChatModel) renderEmptyState(width, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	title := m.emptyStateTitle(width)
	subtitle := emptyStateSubtitleStyle.Width(width).Align(lipgloss.Center).Render("Start a conversation below")
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		subtitle,
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (m ChatModel) emptyStateTitle(width int) string {
	large := strings.TrimSpace(`
  ________          __  __________  ____
 / ____/ /_  ____ _/ /_/_  __/ / / /  _/
/ /   / __ \/ __ '/ __/ / / / / / // /
/ /___/ / / / /_/ / /_  / / / /_/ // /
\____/_/ /_/\__,_/\__/ /_/  \____/___/
`)
	compact := strings.TrimSpace(`
  chatTUI
`)

	title := large
	if width < lipgloss.Width(large) {
		title = compact
	}

	return emptyStateTitleStyle.Width(width).Align(lipgloss.Center).Render(title)
}

func (m ChatModel) transcriptPaneWidth() int {
	paneInnerWidth := m.viewport.Width()
	if paneInnerWidth < 1 {
		paneInnerWidth = m.width - appStyle.GetHorizontalFrameSize() - paneStyle.GetHorizontalFrameSize()
	}
	if paneInnerWidth < 1 {
		paneInnerWidth = 1
	}

	return paneInnerWidth
}

func (m ChatModel) conversationWidth(paneWidth int) int {
	conversationWidth := paneWidth
	if conversationWidth > 84 {
		conversationWidth = 84
	}
	maxAvailableWidth := paneWidth - 2
	if maxAvailableWidth < 1 {
		maxAvailableWidth = 1
	}
	if conversationWidth > maxAvailableWidth {
		conversationWidth = maxAvailableWidth
	}
	if conversationWidth < 1 {
		conversationWidth = 1
	}

	return conversationWidth
}

func (m ChatModel) conversationLaneWidth(conversationWidth int) int {
	conversationLaneWidth := conversationWidth
	if conversationLaneWidth < conversationWidth/2 {
		conversationLaneWidth = conversationWidth / 2
	}
	if conversationLaneWidth < 1 {
		conversationLaneWidth = 1
	}

	return conversationLaneWidth
}

func (m ChatModel) assistantBubbleWidth(laneWidth int) int {
	bubbleWidth := laneWidth * 2 / 3
	if bubbleWidth > 72 {
		bubbleWidth = 72
	}
	maxAvailableWidth := laneWidth
	if maxAvailableWidth < 1 {
		maxAvailableWidth = 1
	}
	if bubbleWidth > maxAvailableWidth {
		bubbleWidth = maxAvailableWidth
	}
	if bubbleWidth < 1 {
		bubbleWidth = 1
	}

	return bubbleWidth
}

func (m ChatModel) userBubbleWidth(laneWidth int) int {
	bubbleWidth := laneWidth * 3 / 5
	if bubbleWidth > 64 {
		bubbleWidth = 64
	}
	maxAvailableWidth := laneWidth
	if maxAvailableWidth < 1 {
		maxAvailableWidth = 1
	}
	if bubbleWidth > maxAvailableWidth {
		bubbleWidth = maxAvailableWidth
	}
	if bubbleWidth < 1 {
		bubbleWidth = 1
	}

	return bubbleWidth
}

// FOR FUTURE USE
// func (m Model) renderComposerLabel(width int) string {
// 	style := composerLabelBlurredStyle
// 	if m.textinput.Focused() {
// 		style = composerLabelFocusedStyle
// 	}

// 	return style.Width(width).Render("Message")
// }

// func (m Model) renderComposerHint(width int) string {
// 	return composerHintStyle.Width(width).Render("Enter to send")
// }

func (m ChatModel) renderLayoutSections(width int) layoutSections {
	return layoutSections{
		header:   m.renderHeader(width),
		status:   m.renderStatus(width),
		footer:   m.renderFooter(width),
		composer: m.renderComposer(width),
	}
}

package main

import (
	"log"
	"strings"

	"charm.land/bubbles/v2/spinner"
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

type ChatRequest struct {
	Messages []ChatMessage
}

type ChatResponse struct {
	Response string
}

type Model struct {
	spinner          spinner.Model
	viewport         viewport.Model
	textinput        textinput.Model
	messages         []ChatMessage
	input            string
	pending          bool
	err              error
	width            int
	height           int
	cursor           int
	client           ChatClient
	chatClientConfig ChatClientConfig
	chatrequest      ChatRequest
	chatresponse     ChatResponse
}

type layoutSections struct {
	header   string
	status   string
	composer string
	footer   string
}

func newModel(config ChatClientConfig) Model {
	vp := viewport.New(
		viewport.WithWidth(80),
		viewport.WithHeight(20),
	)

	s := spinner.New()
	s.Spinner = spinner.Points

	ti := textinput.New()
	ti.Placeholder = "Ask anything"
	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(20)

	return Model{
		spinner:          s,
		viewport:         vp,
		textinput:        ti,
		pending:          false,
		messages:         []ChatMessage{},
		chatClientConfig: config,
		client:           newChatClient(config),
	}
}

func (m Model) renderMessages() string {
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

func sendMessages(m Model) tea.Cmd {
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

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {

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
		composerInnerWidth := sectionWidth - m.currentComposerStyle().GetHorizontalFrameSize()

		if paneInnerWidth < 1 {
			paneInnerWidth = 1
		}
		if composerInnerWidth < 1 {
			composerInnerWidth = 1
		}

		m.viewport.SetWidth(paneInnerWidth)
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

		switch msg.String() {

		case "ctrl+c", "esc":
			return m, tea.Quit

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

func (m Model) View() tea.View {
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
	v.Cursor = c
	return v
}

func (m Model) renderHeader(width int) string {
	innerWidth := width - headerStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	return headerStyle.Width(innerWidth).Render("Active Model: " + m.chatClientConfig.Model)
}

func (m Model) renderStatus(width int) string {
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

func (m Model) renderPane(width int) string {
	innerWidth := width - paneStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	content := m.viewport.View()
	if len(m.messages) == 0 {
		content = m.renderEmptyState(innerWidth, m.viewport.Height())
	}
	return paneStyle.Width(innerWidth).Render(content)
}

func (m Model) renderComposer(width int) string {
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

func (m Model) renderFooter(width int) string {
	innerWidth := width - footerStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	return footerStyle.Width(innerWidth).Render("Enter send | Ctrl+C or esc to exit")
}

func (m Model) currentComposerStyle() lipgloss.Style {
	if m.textinput.Focused() {
		return composerFocusedStyle
	}

	return composerBlurredStyle
}

func (m Model) renderEmptyState(width, height int) string {
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

func (m Model) emptyStateTitle(width int) string {
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

func (m Model) transcriptPaneWidth() int {
	paneInnerWidth := m.viewport.Width()
	if paneInnerWidth < 1 {
		paneInnerWidth = m.width - appStyle.GetHorizontalFrameSize() - paneStyle.GetHorizontalFrameSize()
	}
	if paneInnerWidth < 1 {
		paneInnerWidth = 1
	}

	return paneInnerWidth
}

func (m Model) conversationWidth(paneWidth int) int {
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

func (m Model) conversationLaneWidth(conversationWidth int) int {
	conversationLaneWidth := conversationWidth
	if conversationLaneWidth < conversationWidth/2 {
		conversationLaneWidth = conversationWidth / 2
	}
	if conversationLaneWidth < 1 {
		conversationLaneWidth = 1
	}

	return conversationLaneWidth
}

func (m Model) assistantBubbleWidth(laneWidth int) int {
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

func (m Model) userBubbleWidth(laneWidth int) int {
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

func (m Model) renderLayoutSections(width int) layoutSections {
	return layoutSections{
		header:   m.renderHeader(width),
		status:   m.renderStatus(width),
		footer:   m.renderFooter(width),
		composer: m.renderComposer(width),
	}
}

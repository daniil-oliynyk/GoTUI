package main

import "charm.land/lipgloss/v2"

var botStyle = lipgloss.NewStyle().
	Padding(0, 1).
	Background(lipgloss.Color("238")).
	Foreground(lipgloss.Color("255"))

var userStyle = lipgloss.NewStyle().
	Padding(0, 1).
	Background(lipgloss.Color("62")).
	Foreground(lipgloss.Color("230"))

var appStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(0, 1)

var headerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("230")).
	Background(lipgloss.Color("62")).
	Padding(0, 1)

var paneStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240"))

var emptyStateTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("69"))

var emptyStateSubtitleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("244"))

var composerFocusedStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("69")).
	Padding(1, 1)

var composerBlurredStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Background(lipgloss.Color("235")).
	Padding(1, 1)

var composerLabelFocusedStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("117")).
	MarginBottom(1)

var composerLabelBlurredStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("244")).
	MarginBottom(1)

var composerHintStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241")).
	MarginTop(1)

var composerInputStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("226")).
	Foreground(lipgloss.Color("235"))

var statusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("244"))

var footerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241"))

package main

import "charm.land/lipgloss/v2"

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

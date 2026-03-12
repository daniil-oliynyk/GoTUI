package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {

	m := newModel()
	p := tea.NewProgram(m)

	_, err := p.Run()

	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}

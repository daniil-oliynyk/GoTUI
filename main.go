package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	fd, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Println("Error creating log file:", err)
		os.Exit(1)
	}
	defer fd.Close()

	config := AppConfig{
		APIKey: "temp key",
		Model:  "temp model",
	}
	m := newModel(config)
	p := tea.NewProgram(m)

	_, err = p.Run()

	if err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}

}

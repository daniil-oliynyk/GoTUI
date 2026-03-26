package main

import (
	"context"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/caarlos0/env/v11"
)

type ChatClientConfig struct {
	APIKey string `env:"API_KEY"`
	Model  string
}

func main() {
	fd, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Println("Error creating log file:", err)
		os.Exit(1)
	}
	defer fd.Close()

	config := ChatClientConfig{Model: "gpt-5-nano-2025-08-07"}
	err = env.Parse(&config)
	if err != nil {
		log.Println("Error parsing config:", err)
		os.Exit(1)
	}
	log.Println("Config loaded")

	sqliteDB, err := NewSQLiteDB("chat.db")
	if err != nil {
		log.Println("Error opening sqlite db:", err)
		os.Exit(1)
	}
	defer sqliteDB.Close()

	migrator := NewSQLiteMigrator(sqliteDB.DB())
	if err := migrator.Migrate(context.Background()); err != nil {
		log.Println("Error running sqlite migrations:", err)
		os.Exit(1)
	}

	sessionStore := NewSQLiteSessionStore(sqliteDB)
	messageStore := NewSQLiteMessageStore(sqliteDB)
	history := NewChatHistoryImpl(sessionStore, messageStore)

	m := newAppModel(config, history)
	p := tea.NewProgram(m)
	_, err = p.Run()

	if err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}

}

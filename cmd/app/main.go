package main

import (
	"context"
	"log"
	"os"

	"gotui/internal/app"
	"gotui/internal/config"
	"gotui/internal/store"

	tea "charm.land/bubbletea/v2"
	"github.com/caarlos0/env/v11"
)

func main() {
	fd, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Println("Error creating log file:", err)
		os.Exit(1)
	}
	defer fd.Close()

	cfg := config.ChatClientConfig{Model: "gpt-5-nano-2025-08-07"}
	err = env.Parse(&cfg)
	if err != nil {
		log.Println("Error parsing config:", err)
		os.Exit(1)
	}
	log.Println("Config loaded")

	sqliteDB, err := store.NewSQLiteDB("chat.db")
	if err != nil {
		log.Println("Error opening sqlite db:", err)
		os.Exit(1)
	}
	defer sqliteDB.Close()

	migrator := store.NewSQLiteMigrator(sqliteDB.DB())
	if err := migrator.Migrate(context.Background()); err != nil {
		log.Println("Error running sqlite migrations:", err)
		os.Exit(1)
	}

	sessionStore := store.NewSQLiteSessionStore(sqliteDB)
	messageStore := store.NewSQLiteMessageStore(sqliteDB)
	history := store.NewChatHistoryImpl(sessionStore, messageStore)

	m := app.NewModel(cfg, history)
	p := tea.NewProgram(m)
	_, err = p.Run()

	if err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}

}

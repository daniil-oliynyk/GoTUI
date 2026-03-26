package main

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type SessionItem struct {
	ID        string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Messages  []ChatMessage
}

type SessionStore interface {
	CreateSession(ctx context.Context, title string) (SessionItem, error)
	DeleteSession(ctx context.Context, id string) error
	ListSessions(ctx context.Context) ([]SessionItem, error)
	GetSession(ctx context.Context, id string) (SessionItem, error)
	UpdateSession(ctx context.Context, id string, title string) (SessionItem, error)
}

type MessageStore interface {
	CreateMessage(ctx context.Context, sessionID string, message ChatMessage) (ChatMessage, error)
	ListMessagesBySession(ctx context.Context, sessionID string) ([]ChatMessage, error)
	DeleteMessagesBySession(ctx context.Context, id string) error
}

type ChatHistory interface {
	InitDefaultSession(ctx context.Context) (SessionItem, error)
	LoadSessions(ctx context.Context) ([]SessionItem, error)
	LoadSessionMessages(ctx context.Context, sessionID string) ([]ChatMessage, error)
	AddUserMessage(ctx context.Context, sessionID string, message ChatMessage) (ChatMessage, error)
	AddAssistantMessage(ctx context.Context, sessionID string, message ChatMessage) (ChatMessage, error)
	CreateAndSelectSession(ctx context.Context, title string) (SessionItem, error)
	DeleteSession(ctx context.Context, id string) (next SessionItem, err error)
}

type SQLiteDB struct {
	db *sql.DB
}

func NewSQLiteDB(path string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{db: db}, nil
}

func (s *SQLiteDB) DB() *sql.DB {
	return s.db
}

func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

type SQLiteMigrator struct {
	db *sql.DB
}

func (m *SQLiteMigrator) Migrate(ctx context.Context) error {
	statements := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY_KEY,
			title TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME NULL
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY_KEY,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session_created_at ON messages(session_id, created_at ASC);`,
	}

	return nil
}

type SQLiteSessionStore struct {
	db *SQLiteDB
}

func NewSQLiteSessionStore(db *SQLiteDB) *SQLiteSessionStore {
	return &SQLiteSessionStore{db: db}
}

func (s *SQLiteSessionStore) CreateSession(ctx context.Context, title string) (SessionItem, error) {
	return SessionItem{}, nil
}

func (s *SQLiteSessionStore) DeleteSession(ctx context.Context, id string) error {
	return nil
}

func (s *SQLiteSessionStore) ListSessions(ctx context.Context) ([]SessionItem, error) {
	return nil, nil
}

func (s *SQLiteSessionStore) GetSession(ctx context.Context, id string) (SessionItem, error) {
	return SessionItem{}, nil
}

func (s *SQLiteSessionStore) UpdateSession(ctx context.Context, id string, title string) (SessionItem, error) {
	return SessionItem{}, nil
}

type SQLiteMessageStore struct {
	db *SQLiteDB
}

func NewSQLiteMessageStore(db *SQLiteDB) *SQLiteMessageStore {
	return &SQLiteMessageStore{db: db}
}

func (s *SQLiteMessageStore) CreateMessage(ctx context.Context, sessionID string, message ChatMessage) (ChatMessage, error) {
	return ChatMessage{}, nil
}

func (s *SQLiteMessageStore) ListMessagesBySession(ctx context.Context, sessionID string) ([]ChatMessage, error) {
	return nil, nil
}

func (s *SQLiteMessageStore) DeleteMessagesBySession(ctx context.Context, id string) error {
	return nil
}

type ChatHistoryImpl struct {
	sessionStore SessionStore
	messageStore MessageStore
}

func NewChatHistoryImpl(sessionStore SessionStore, messageStore MessageStore) *ChatHistoryImpl {
	return &ChatHistoryImpl{
		sessionStore: sessionStore,
		messageStore: messageStore,
	}
}

func (c *ChatHistoryImpl) InitDefaultSession(ctx context.Context) (SessionItem, error) {
	return SessionItem{}, nil
}

func (c *ChatHistoryImpl) LoadSessions(ctx context.Context) ([]SessionItem, error) {
	return nil, nil
}

func (c *ChatHistoryImpl) LoadSessionMessages(ctx context.Context, sessionID string) ([]ChatMessage, error) {
	return nil, nil
}

func (c *ChatHistoryImpl) AddUserMessage(ctx context.Context, sessionID string, message ChatMessage) (ChatMessage, error) {
	return ChatMessage{}, nil
}

func (c *ChatHistoryImpl) AddAssistantMessage(ctx context.Context, sessionID string, message ChatMessage) (ChatMessage, error) {
	return ChatMessage{}, nil
}

func (c *ChatHistoryImpl) CreateAndSelectSession(ctx context.Context, title string) (SessionItem, error) {
	return SessionItem{}, nil
}

func (c *ChatHistoryImpl) DeleteSession(ctx context.Context, id string) (SessionItem, error) {
	return SessionItem{}, nil
}

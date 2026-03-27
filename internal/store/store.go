package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gotui/internal/domain"

	_ "modernc.org/sqlite"
)

type SessionItem = domain.SessionItem
type ChatMessage = domain.ChatMessage
type MessageRole = domain.MessageRole

const (
	MessageRoleUser      = domain.MessageRoleUser
	MessageRoleAssistant = domain.MessageRoleAssistant
)

var ErrNotFound = errors.New("not found")

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
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
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

func NewSQLiteMigrator(db *sql.DB) *SQLiteMigrator {
	return &SQLiteMigrator{db: db}
}

func (m *SQLiteMigrator) Migrate(ctx context.Context) error {
	statements := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME NULL
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session_created_at ON messages(session_id, created_at ASC);`,
	}

	for i, stmt := range statements {
		if _, err := m.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migration statement %d failed: %w", i+1, err)
		}
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
	id := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	now := time.Now().UTC()
	if title == "" {
		title = "Session"
	}

	_, err := s.db.DB().ExecContext(
		ctx,
		`INSERT INTO sessions (id, title, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, NULL)`,
		id,
		title,
		now,
		now,
	)
	if err != nil {
		return SessionItem{}, fmt.Errorf("create session: %w", err)
	}

	return SessionItem{ID: id, Title: title, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *SQLiteSessionStore) DeleteSession(ctx context.Context, id string) error {
	now := time.Now().UTC()
	res, err := s.db.DB().ExecContext(
		ctx,
		`UPDATE sessions SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		now,
		now,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete session rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *SQLiteSessionStore) ListSessions(ctx context.Context) ([]SessionItem, error) {
	rows, err := s.db.DB().QueryContext(
		ctx,
		`SELECT id, title, created_at, updated_at, deleted_at FROM sessions WHERE deleted_at IS NULL ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	sessions := []SessionItem{}
	for rows.Next() {
		var item SessionItem
		var deletedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Title, &item.CreatedAt, &item.UpdatedAt, &deletedAt); err != nil {
			return nil, fmt.Errorf("scan session row: %w", err)
		}
		if deletedAt.Valid {
			item.DeletedAt = &deletedAt.Time
		}
		sessions = append(sessions, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}

	return sessions, nil
}

func (s *SQLiteSessionStore) GetSession(ctx context.Context, id string) (SessionItem, error) {
	var item SessionItem
	var deletedAt sql.NullTime
	err := s.db.DB().QueryRowContext(
		ctx,
		`SELECT id, title, created_at, updated_at, deleted_at FROM sessions WHERE id = ? AND deleted_at IS NULL LIMIT 1`,
		id,
	).Scan(&item.ID, &item.Title, &item.CreatedAt, &item.UpdatedAt, &deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return SessionItem{}, ErrNotFound
	}
	if err != nil {
		return SessionItem{}, fmt.Errorf("get session: %w", err)
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}
	return item, nil
}

func (s *SQLiteSessionStore) UpdateSession(ctx context.Context, id string, title string) (SessionItem, error) {
	now := time.Now().UTC()
	res, err := s.db.DB().ExecContext(
		ctx,
		`UPDATE sessions SET title = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		title,
		now,
		id,
	)
	if err != nil {
		return SessionItem{}, fmt.Errorf("update session: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return SessionItem{}, fmt.Errorf("update session rows affected: %w", err)
	}
	if rows == 0 {
		return SessionItem{}, ErrNotFound
	}

	item, err := s.GetSession(ctx, id)
	if err != nil {
		return SessionItem{}, fmt.Errorf("get updated session: %w", err)
	}
	return item, nil
}

type SQLiteMessageStore struct {
	db *SQLiteDB
}

func NewSQLiteMessageStore(db *SQLiteDB) *SQLiteMessageStore {
	return &SQLiteMessageStore{db: db}
}

func (s *SQLiteMessageStore) CreateMessage(ctx context.Context, sessionID string, message ChatMessage) (ChatMessage, error) {
	if message.ID == "" {
		message.ID = fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.DB().ExecContext(
		ctx,
		`INSERT INTO messages (id, session_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		message.ID,
		sessionID,
		string(message.Role),
		message.Content,
		message.CreatedAt,
	)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("create message: %w", err)
	}

	return message, nil
}

func (s *SQLiteMessageStore) ListMessagesBySession(ctx context.Context, sessionID string) ([]ChatMessage, error) {
	rows, err := s.db.DB().QueryContext(
		ctx,
		`SELECT id, role, content, created_at FROM messages WHERE session_id = ? ORDER BY created_at ASC`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list messages by session: %w", err)
	}
	defer rows.Close()

	messages := []ChatMessage{}
	for rows.Next() {
		var msg ChatMessage
		var role string
		if err := rows.Scan(&msg.ID, &role, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message row: %w", err)
		}
		msg.Role = MessageRole(role)
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	return messages, nil
}

func (s *SQLiteMessageStore) DeleteMessagesBySession(ctx context.Context, id string) error {
	_, err := s.db.DB().ExecContext(ctx, `DELETE FROM messages WHERE session_id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete messages by session: %w", err)
	}
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
	sessions, err := c.sessionStore.ListSessions(ctx)
	if err != nil {
		return SessionItem{}, fmt.Errorf("init default session list: %w", err)
	}
	if len(sessions) == 0 {
		session, err := c.sessionStore.CreateSession(ctx, "Session 1")
		if err != nil {
			return SessionItem{}, fmt.Errorf("init default session create: %w", err)
		}
		return session, nil
	}

	return sessions[0], nil
}

func (c *ChatHistoryImpl) LoadSessions(ctx context.Context) ([]SessionItem, error) {
	sessions, err := c.sessionStore.ListSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("load sessions: %w", err)
	}
	return sessions, nil
}

func (c *ChatHistoryImpl) LoadSessionMessages(ctx context.Context, sessionID string) ([]ChatMessage, error) {
	messages, err := c.messageStore.ListMessagesBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("load session messages: %w", err)
	}
	return messages, nil
}

func (c *ChatHistoryImpl) AddUserMessage(ctx context.Context, sessionID string, message ChatMessage) (ChatMessage, error) {
	message.Role = MessageRoleUser
	persisted, err := c.messageStore.CreateMessage(ctx, sessionID, message)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("add user message: %w", err)
	}
	if err := c.touchSession(ctx, sessionID); err != nil {
		return ChatMessage{}, err
	}
	return persisted, nil
}

func (c *ChatHistoryImpl) AddAssistantMessage(ctx context.Context, sessionID string, message ChatMessage) (ChatMessage, error) {
	message.Role = MessageRoleAssistant
	persisted, err := c.messageStore.CreateMessage(ctx, sessionID, message)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("add assistant message: %w", err)
	}
	if err := c.touchSession(ctx, sessionID); err != nil {
		return ChatMessage{}, err
	}
	return persisted, nil
}

func (c *ChatHistoryImpl) CreateAndSelectSession(ctx context.Context, title string) (SessionItem, error) {
	session, err := c.sessionStore.CreateSession(ctx, title)
	if err != nil {
		return SessionItem{}, fmt.Errorf("create and select session: %w", err)
	}
	return session, nil
}

func (c *ChatHistoryImpl) DeleteSession(ctx context.Context, id string) (SessionItem, error) {
	sessions, err := c.sessionStore.ListSessions(ctx)
	if err != nil {
		return SessionItem{}, fmt.Errorf("delete session list sessions: %w", err)
	}
	if len(sessions) <= 1 {
		return SessionItem{}, fmt.Errorf("cannot delete last session")
	}

	targetIndex := -1
	for i, session := range sessions {
		if session.ID == id {
			targetIndex = i
			break
		}
	}
	if targetIndex == -1 {
		return SessionItem{}, ErrNotFound
	}

	if err := c.messageStore.DeleteMessagesBySession(ctx, id); err != nil {
		return SessionItem{}, fmt.Errorf("delete session messages: %w", err)
	}
	if err := c.sessionStore.DeleteSession(ctx, id); err != nil {
		return SessionItem{}, fmt.Errorf("delete session row: %w", err)
	}

	sessions, err = c.sessionStore.ListSessions(ctx)
	if err != nil {
		return SessionItem{}, fmt.Errorf("delete session reload sessions: %w", err)
	}
	if len(sessions) == 0 {
		session, err := c.sessionStore.CreateSession(ctx, "Session 1")
		if err != nil {
			return SessionItem{}, fmt.Errorf("delete session recreate default: %w", err)
		}
		return session, nil
	}

	if targetIndex >= len(sessions) {
		targetIndex = len(sessions) - 1
	}
	if targetIndex < 0 {
		targetIndex = 0
	}

	return sessions[targetIndex], nil
}

func (c *ChatHistoryImpl) touchSession(ctx context.Context, sessionID string) error {
	session, err := c.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("touch session get: %w", err)
	}
	if _, err := c.sessionStore.UpdateSession(ctx, sessionID, session.Title); err != nil {
		return fmt.Errorf("touch session update: %w", err)
	}
	return nil
}

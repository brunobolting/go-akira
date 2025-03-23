package session

import (
	"akira/internal/entity"
	"database/sql"
	"encoding/json"
	"time"
)

var _ entity.SessionRepository = (*SessionSqliteRepository)(nil)

type SessionSqliteRepository struct {
	db *sql.DB
}

func NewSessionSqliteRepository(db *sql.DB) *SessionSqliteRepository {
	return &SessionSqliteRepository{db: db}
}

func (r *SessionSqliteRepository) scanSessionRow(row entity.Rowscan) (*entity.Session, error) {
	var session entity.Session
	var data []byte
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&data,
		&session.CreatedAt,
		&session.ExpiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	session.Data = make(map[string]any)
	if err := json.Unmarshal(data, &session.Data); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionSqliteRepository) FindSession(id string) (*entity.Session, error) {
	stmt, err := r.db.Prepare("SELECT id, user_id, data, created_at, expires_at FROM sessions WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(id)
	return r.scanSessionRow(row)
}

func (r *SessionSqliteRepository) CreateSession(session *entity.Session) error {
	data, err := json.Marshal(session.Data)
	if err != nil {
		return err
	}
	stmt, err := r.db.Prepare("INSERT INTO sessions (id, user_id, data, created_at, expires_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(session.ID, session.UserID, data, session.CreatedAt, session.ExpiresAt)
	return err
}

func (r *SessionSqliteRepository) DeleteSession(id string) error {
	stmt, err := r.db.Prepare("DELETE FROM sessions WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	return err
}

func (r *SessionSqliteRepository) GC() error {
	stmt, err := r.db.Prepare("DELETE FROM sessions WHERE expires_at < ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(time.Now().UTC())
	return err
}

func (r *SessionSqliteRepository) GetExpiredSessions() ([]entity.Session, error) {
	stmt, err := r.db.Prepare("SELECT id, user_id, data, created_at, expires_at FROM sessions WHERE expires_at < ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []entity.Session
	for rows.Next() {
		s, err := r.scanSessionRow(rows)
		if err != nil {
			return nil, err
		}
		if s != nil {
			sessions = append(sessions, *s)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

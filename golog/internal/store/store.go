// Package store keeps projects and logs in a single SQLite file.
package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golog/internal/model"

	_ "modernc.org/sqlite"
)

// ErrNotFound means the project or log does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict means the log was changed since the caller read it.
var ErrConflict = errors.New("the log was changed elsewhere")

// Store is the database. It is safe for concurrent use.
type Store struct{ db *sql.DB }

// migrations are applied in order; PRAGMA user_version records how many have run.
var migrations = []string{
	`CREATE TABLE projects (
		id      TEXT PRIMARY KEY,
		name    TEXT NOT NULL,
		created TEXT NOT NULL,
		updated TEXT NOT NULL
	);
	CREATE TABLE logs (
		id         TEXT PRIMARY KEY,
		project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		name       TEXT NOT NULL,
		revision   INTEGER NOT NULL,
		content    TEXT NOT NULL,
		bed_count  INTEGER NOT NULL,
		created    TEXT NOT NULL,
		updated    TEXT NOT NULL
	);
	CREATE INDEX logs_project ON logs(project_id);`,
}

// Open opens (or creates) the database file at path and brings its schema up to date.
func Open(path string) (*Store, error) {
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite allows one writer; one connection avoids "database is locked"
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("preparing database %s: %w", path, err)
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	var version int
	if err := s.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	if version > len(migrations) {
		return fmt.Errorf("the database was made by a newer GoLog (schema %d)", version)
	}
	for i := version; i < len(migrations); i++ {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(migrations[i]); err != nil {
			tx.Rollback()
			return err
		}
		if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, i+1)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// NewID returns a random 16-character id.
func NewID() string {
	b := make([]byte, 10)
	rand.Read(b)
	return strings.ToLower(base32.StdEncoding.EncodeToString(b))
}

func now() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }

func ts(t time.Time) string { return t.Format(time.RFC3339Nano) }

func parseTS(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}

// Projects lists all projects, most recently changed first.
func (s *Store) Projects(ctx context.Context) ([]model.ProjectSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.updated, COUNT(l.id)
		FROM projects p LEFT JOIN logs l ON l.project_id = p.id
		GROUP BY p.id ORDER BY p.updated DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []model.ProjectSummary{}
	for rows.Next() {
		var p model.ProjectSummary
		var updated string
		if err := rows.Scan(&p.ID, &p.Name, &updated, &p.LogCount); err != nil {
			return nil, err
		}
		p.Updated = parseTS(updated)
		list = append(list, p)
	}
	return list, rows.Err()
}

// CreateProject adds a project.
func (s *Store) CreateProject(ctx context.Context, name string) (*model.Project, error) {
	t := now()
	p := &model.Project{ID: NewID(), Name: name, Created: t, Updated: t, Logs: []model.LogSummary{}}
	_, err := s.db.ExecContext(ctx, `INSERT INTO projects (id, name, created, updated) VALUES (?, ?, ?, ?)`,
		p.ID, p.Name, ts(t), ts(t))
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Project returns a project with its list of logs.
func (s *Store) Project(ctx context.Context, id string) (*model.Project, error) {
	var p model.Project
	var created, updated string
	err := s.db.QueryRowContext(ctx, `SELECT id, name, created, updated FROM projects WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Created, p.Updated = parseTS(created), parseTS(updated)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, updated, bed_count FROM logs WHERE project_id = ? ORDER BY name COLLATE NOCASE`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	p.Logs = []model.LogSummary{}
	for rows.Next() {
		var l model.LogSummary
		var lu string
		if err := rows.Scan(&l.ID, &l.Name, &lu, &l.BedCount); err != nil {
			return nil, err
		}
		l.Updated = parseTS(lu)
		p.Logs = append(p.Logs, l)
	}
	return &p, rows.Err()
}

// RenameProject changes a project's name.
func (s *Store) RenameProject(ctx context.Context, id, name string) error {
	return s.expectOne(s.db.ExecContext(ctx, `UPDATE projects SET name = ?, updated = ? WHERE id = ?`, name, ts(now()), id))
}

// DeleteProject removes a project and all its logs.
func (s *Store) DeleteProject(ctx context.Context, id string) error {
	return s.expectOne(s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id))
}

// CreateLog adds a log to a project: empty, or with the given (already validated) content.
func (s *Store) CreateLog(ctx context.Context, projectID, name string, content ...model.LogContent) (*model.Log, error) {
	t := now()
	l := &model.Log{ID: NewID(), ProjectID: projectID, Name: name, Revision: 1, Created: t, Updated: t, LogContent: model.NewLogContent()}
	if len(content) > 0 {
		l.LogContent = content[0]
	}
	data, _ := json.Marshal(l.LogContent)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := s.expectOne(tx.ExecContext(ctx, `UPDATE projects SET updated = ? WHERE id = ?`, ts(t), projectID)); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO logs (id, project_id, name, revision, content, bed_count, created, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, l.ID, projectID, name, l.Revision, string(data), len(l.Beds), ts(t), ts(t)); err != nil {
		return nil, err
	}
	return l, tx.Commit()
}

// Log returns a log with its beds.
func (s *Store) Log(ctx context.Context, id string) (*model.Log, error) {
	var l model.Log
	var content, created, updated string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, revision, content, created, updated FROM logs WHERE id = ?`, id).
		Scan(&l.ID, &l.ProjectID, &l.Name, &l.Revision, &content, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(content), &l.LogContent); err != nil {
		return nil, fmt.Errorf("log %s is damaged: %w", id, err)
	}
	l.LogContent.Normalize()
	l.Created, l.Updated = parseTS(created), parseTS(updated)
	return &l, nil
}

// UpdateLog saves a log's name and content. It fails with ErrConflict unless
// l.Revision is the revision currently stored; on success l gets the new revision.
func (s *Store) UpdateLog(ctx context.Context, l *model.Log) error {
	content, err := json.Marshal(l.LogContent)
	if err != nil {
		return err
	}
	t := now()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var projectID string
	var revision int64
	err = tx.QueryRowContext(ctx, `SELECT project_id, revision FROM logs WHERE id = ?`, l.ID).Scan(&projectID, &revision)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if revision != l.Revision {
		return ErrConflict
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE logs SET name = ?, revision = revision + 1, content = ?, bed_count = ?, updated = ? WHERE id = ?`,
		l.Name, string(content), len(l.Beds), ts(t), l.ID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE projects SET updated = ? WHERE id = ?`, ts(t), projectID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	l.ProjectID, l.Revision, l.Updated = projectID, revision+1, t
	return nil
}

// DeleteLog removes a log.
func (s *Store) DeleteLog(ctx context.Context, id string) error {
	return s.expectOne(s.db.ExecContext(ctx, `DELETE FROM logs WHERE id = ?`, id))
}

func (s *Store) expectOne(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

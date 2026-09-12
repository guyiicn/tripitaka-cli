package state

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Progress struct {
	SutraID string
	Juan    string
	Title   string
	Char    int
	Context string
	Updated int64
}

type Bookmark struct {
	ID      int64
	SutraID string
	Juan    string
	Title   string
	Char    int
	Snippet string
	Created int64
}

type Store struct{ db *sql.DB }

func DefaultPath() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "tripitaka-cli", "state.db"), nil
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	_, err = db.Exec(`
PRAGMA journal_mode=WAL;
CREATE TABLE IF NOT EXISTS progress(
  sutra_id TEXT NOT NULL, juan TEXT NOT NULL, title TEXT NOT NULL,
  char_index INTEGER NOT NULL, context TEXT NOT NULL DEFAULT '',
  updated_at INTEGER NOT NULL, PRIMARY KEY(sutra_id, juan)
);
CREATE TABLE IF NOT EXISTS bookmarks(
  id INTEGER PRIMARY KEY AUTOINCREMENT, sutra_id TEXT NOT NULL, juan TEXT NOT NULL,
  title TEXT NOT NULL, char_index INTEGER NOT NULL, snippet TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS bookmarks_recent ON bookmarks(created_at DESC);
`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) SaveProgress(p Progress) error {
	p.Updated = time.Now().UnixNano()
	_, err := s.db.Exec(`
INSERT INTO progress(sutra_id,juan,title,char_index,context,updated_at) VALUES(?,?,?,?,?,?)
ON CONFLICT(sutra_id,juan) DO UPDATE SET title=excluded.title,char_index=excluded.char_index,
context=excluded.context,updated_at=excluded.updated_at`,
		p.SutraID, p.Juan, p.Title, p.Char, p.Context, p.Updated)
	return err
}

func (s *Store) Progress(id, juan string) (Progress, error) {
	var p Progress
	err := s.db.QueryRow(`SELECT sutra_id,juan,title,char_index,context,updated_at
FROM progress WHERE sutra_id=? AND juan=?`, id, juan).
		Scan(&p.SutraID, &p.Juan, &p.Title, &p.Char, &p.Context, &p.Updated)
	return p, err
}

func (s *Store) LatestProgress(id string) (Progress, error) {
	var p Progress
	err := s.db.QueryRow(`SELECT sutra_id,juan,title,char_index,context,updated_at
FROM progress WHERE sutra_id=? ORDER BY updated_at DESC LIMIT 1`, id).
		Scan(&p.SutraID, &p.Juan, &p.Title, &p.Char, &p.Context, &p.Updated)
	return p, err
}

func (s *Store) Recent(limit int) ([]Progress, error) {
	rows, err := s.db.Query(`SELECT sutra_id,juan,title,char_index,context,updated_at
FROM progress ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Progress
	for rows.Next() {
		var p Progress
		if err := rows.Scan(&p.SutraID, &p.Juan, &p.Title, &p.Char, &p.Context, &p.Updated); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) AddBookmark(b Bookmark) error {
	b.Created = time.Now().UnixNano()
	_, err := s.db.Exec(`INSERT INTO bookmarks(sutra_id,juan,title,char_index,snippet,created_at)
VALUES(?,?,?,?,?,?)`, b.SutraID, b.Juan, b.Title, b.Char, b.Snippet, b.Created)
	return err
}

func (s *Store) Bookmarks(limit int) ([]Bookmark, error) {
	rows, err := s.db.Query(`SELECT id,sutra_id,juan,title,char_index,snippet,created_at
FROM bookmarks ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Bookmark
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.SutraID, &b.Juan, &b.Title, &b.Char, &b.Snippet, &b.Created); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) DeleteBookmark(id int64) error {
	res, err := s.db.Exec("DELETE FROM bookmarks WHERE id=?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("bookmark not found")
	}
	return nil
}

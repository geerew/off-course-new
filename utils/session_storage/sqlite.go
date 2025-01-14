package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Storage interface that is implemented by storage providers
type Storage struct {
	db         *sql.DB
	table      string
	gcInterval time.Duration
	done       chan struct{}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new storage
func NewSqlite(db *sql.DB, table string) *Storage {
	store := &Storage{
		db:         db,
		table:      table,
		gcInterval: 10 * time.Second,
		done:       make(chan struct{}),
	}

	// Start garbage collector
	go store.gcTicker()

	return store
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get value by key
func (s *Storage) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, nil
	}

	query := fmt.Sprintf("SELECT data, expires FROM %s WHERE id=?", s.table)
	row := s.db.QueryRow(query, key)

	// Add db response to data
	var (
		data       = []byte{}
		exp  int64 = 0
	)

	if err := row.Scan(&data, &exp); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// If the expiration time has already passed, then return nil
	if exp != 0 && exp <= time.Now().Unix() {
		return nil, nil
	}

	return data, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Set key with value
func (s *Storage) Set(key string, data []byte, exp time.Duration) error {
	if key == "" || len(data) <= 0 {
		return nil
	}

	var expSeconds int64
	if exp != 0 {
		expSeconds = time.Now().Add(exp).Unix()
	}

	query := fmt.Sprintf("INSERT OR REPLACE INTO %s (id, data, expires) VALUES (?,?,?)", s.table)
	_, err := s.db.Exec(query, key, data, expSeconds)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes an entry by ID (key)
func (s *Storage) Delete(key string) error {
	if key == "" {
		return nil
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id=?", s.table)
	_, err := s.db.Exec(query, key)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Reset resets all entries, including unexpired
func (s *Storage) Reset() error {
	query := fmt.Sprintf("DELETE FROM %s", s.table)
	_, err := s.db.Exec(query)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Close closes the database
func (s *Storage) Close() error {
	s.done <- struct{}{}
	return s.db.Close()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Conn returns the database client
func (s *Storage) Conn() *sql.DB {
	return s.db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// gcTicker starts the gc ticker
func (s *Storage) gcTicker() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case t := <-ticker.C:
			query := fmt.Sprintf("DELETE FROM %s WHERE expires <= ? AND expires != 0", s.table)
			_, _ = s.db.Exec(query, t.Unix())
		}
	}
}

// SQLite3 FIFO (First In First Out) storage.
package sqlite3

import (
	"bytes"
	"colly/storage"
	"database/sql"
	"io"
)

// ------------------------------------------------------------------------

type stgFIFO struct {
	s *stgBase
}

// ------------------------------------------------------------------------

const defaultFIFOTableName = "fifo"

// ------------------------------------------------------------------------

var (
	cmdFIFO = map[string]string{
		"create":      `CREATE TABLE IF NOT EXISTS "<table>" ("id" INTEGER PRIMARY KEY AUTOINCREMENT, "thread" INTEGER NOT NULL, "data" BLOB)`,
		"drop":        `DROP TABLE IF EXISTS "<table>"`,
		"trim_thread": `DELETE FROM "<table>" WHERE "thread" = ?`,
		"trim":        `DELETE FROM "<table>"`,
		"insert":      `INSERT INTO "<table>" ("thread", "data") VALUES (?, ?)`,
		"select":      `SELECT "data" FROM "<table>" WHERE "id" = (SELECT MIN("id") FROM "<table>" WHERE "thread" = ?)`,
		"pop":         `DELETE FROM "<table>" WHERE "id" = (SELECT MIN("id") FROM "<table>" WHERE "thread" = ?) RETURNING "data"`,
		"multipop":    `DELETE FROM "<table>" WHERE "thread" = ? ORDER BY "id" ASC LIMIT ? RETURNING "data"`,
		"count":       `SELECT COUNT(*) FROM "<table>" WHERE "thread" = ?`,
	}
)

// ------------------------------------------------------------------------

// NewFIFOStorage returns a pointer to a newly created SQLite3 FIFO storage.
func NewFIFOStorage(path string, table string, keepData bool) (*stgFIFO, error) {
	cfg := config{
		table:       setTable(table, defaultFIFOTableName),
		dropOnClose: false,
		clearOnOpen: !keepData,
	}

	s, err := NewBaseStorage(path, &cfg, cmdFIFO)
	if err != nil {
		return nil, err
	}

	return &stgFIFO{
		s: s,
	}, nil
}

// ------------------------------------------------------------------------

// Close closes the SQLite3 FIFO storage.
func (s *stgFIFO) Close() error {
	return s.s.Close()
}

// ------------------------------------------------------------------------

// Clear removes all entries from the SQLite3 FIFO storage.
func (s *stgFIFO) Clear(ids ...uint32) error {
	if len(ids) == 0 {
		return s.s.Clear()
	}

	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	for _, id := range ids {
		err := s.s.Cmd("trim_thread", id)
		if err != nil {
			return err
		}
	}

	return nil
}

// ------------------------------------------------------------------------

// Capacity returns the maximum number of items that can be stored in the FIFO storage.
func (s *stgFIFO) Capacity() uint {
	return 1000000000
}

// ------------------------------------------------------------------------

// Len returns the number of hosts in the SQLite3 FIFO storage.
func (s *stgFIFO) Len(id uint32) (uint, error) {
	return s.s.Len(id)
}

// ------------------------------------------------------------------------

// Push inserts an item into the SQLite3 FIFO storage.
func (s *stgFIFO) Push(id uint32, item io.Reader) error {
	data, err := io.ReadAll(item)
	if err != nil {
		return err
	}

	s.s.lock.Lock()
	_, err = s.s.stmts["insert"].Exec(id, data)
	s.s.lock.Unlock()

	return err
}

// ------------------------------------------------------------------------

// Pop pops the oldest item from the FIFO storage or returns error if the storage is empty.
func (s *stgFIFO) Pop(id uint32) (io.Reader, error) {
	var data = []byte{}

	s.s.lock.Lock()
	err := s.s.stmts["pop"].QueryRow(id).Scan(&data)
	s.s.lock.Unlock()
	if err != nil {
		if err == sql.ErrNoRows {
			err = storage.ErrStorageEmpty
		}

		return nil, err
	}

	return bytes.NewReader(data), nil
}

// ------------------------------------------------------------------------

// MultiPop pops maximum n of the oldest items from the FIFO storage
// or returns error if the storage is empty.
func (s *stgFIFO) MultiPop(id uint32, n uint) ([]io.Reader, error) {
	if n < 1 {
		return nil, storage.ErrInvalidNumber
	}

	s.s.lock.Lock()
	rows, err := s.s.stmts["multipop"].Query(id, n)
	s.s.lock.Unlock()
	if err != nil {
		if err == sql.ErrNoRows {
			err = storage.ErrStorageEmpty
		}

		return nil, err
	}
	defer rows.Close()

	var items = []io.Reader{}
	for rows.Next() {
		var data = []byte{}
		err = rows.Scan(&data)
		if err != nil {
			return items, err
		}

		items = append(items, bytes.NewReader(data))
	}

	return items, nil
}

// ------------------------------------------------------------------------

// Peek returns the oldest item from the FIFO storage without removing it.
func (s *stgFIFO) Peek(id uint32) (io.Reader, error) {
	var data = []byte{}

	s.s.lock.Lock()
	err := s.s.stmts["select"].QueryRow(id).Scan(&data)
	s.s.lock.Unlock()
	if err != nil {
		if err == sql.ErrNoRows {
			err = storage.ErrStorageEmpty
		}

		return nil, err
	}

	return bytes.NewReader(data), nil
}

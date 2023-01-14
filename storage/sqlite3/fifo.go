// SQLite3 FIFO (First In First Out) storage.
package sqlite3

import (
	"colly/storage"
	"database/sql"
	"net/url"
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
		"create": `CREATE TABLE IF NOT EXISTS "<table>" ("id" INTEGER PRIMARY KEY AUTOINCREMENT, "data" BLOB)`,
		"drop":   `DROP TABLE IF EXISTS "<table>"`,
		"trim":   `DELETE FROM "<table>"`,
		"insert": `INSERT INTO "<table>" (data) VALUES (?)`,
		"select": `SELECT "data" FROM "<table>" WHERE "id" = (SELECT MIN("id") FROM "<table>")`,
		"delete": `DELETE FROM "<table>" WHERE "id" = (SELECT MIN("id") FROM "<table>") RETURNING "data"`,
		"count":  `SELECT COUNT(*) FROM "<table>"`,
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
func (s *stgFIFO) Clear() error {
	return s.s.Clear()
}

// ------------------------------------------------------------------------

// Len returns the number of hosts in the SQLite3 FIFO storage.
func (s *stgFIFO) Len() (uint, error) {
	return s.s.Len()
}

// ------------------------------------------------------------------------

// Push inserts an item into the SQLite3 FIFO storage.
func (s *stgFIFO) Push(item []byte) error {
	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	_, err := s.s.stmts["insert"].Exec(item)

	return err
}

// ------------------------------------------------------------------------

// Pop pops the oldest item from the FIFO storage or returns error if the storage is empty.
func (s *stgFIFO) Pop(u *url.URL) ([]byte, error) {
	var item = []byte{}

	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	if err := s.s.stmts["delete"].QueryRow(u.Host).Scan(&item); err != nil {
		if err == sql.ErrNoRows {
			err = storage.ErrStorageEmpty
		}

		return nil, err
	}

	return item, nil
}

// ------------------------------------------------------------------------

// Peek returns the oldest item from the FIFO storage without removing it.
func (s *stgFIFO) Peek() ([]byte, error) {
	var item = []byte{}

	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	if err := s.s.stmts["select"].QueryRow().Scan(&item); err != nil {
		if err == sql.ErrNoRows {
			err = storage.ErrStorageEmpty
		}

		return nil, err
	}

	return item, nil
}

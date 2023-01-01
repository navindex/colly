package sqlite3

// ------------------------------------------------------------------------

type stgVisit struct {
	s *stgBase
}

// ------------------------------------------------------------------------

const defaultVisitTableName = "visited_requests"

// ------------------------------------------------------------------------

var (
	cmdVisit = map[string]string{
		"create": `CREATE TABLE IF NOT EXISTS "<table>" ("id" INTEGER PRIMARY KEY NOT NULL)`,
		"drop":   `DROP TABLE IF EXISTS "<table>"`,
		"trim":   `DELETE FROM "<table>"`,
		"insert": `INSERT INTO "<table>" ("id") VALUES (?) ON CONFLICT("id") DO NOTHING`,
		"select": `SELECT EXISTS(SELECT 1 FROM "<table>" WHERE "id" = ?)`,
		"delete": `DELETE FROM "<table>" WHERE "id" = ?`,
		"count":  `SELECT COUNT(*) FROM "<table>"`,
	}
)

// ------------------------------------------------------------------------

// NewVisitStorage returns a pointer to a newly created SQLite3 visit storage.
func NewVisitStorage(path string, table string, keepData bool) (*stgVisit, error) {
	cfg := config{
		table:       setTable(table, defaultVisitTableName),
		dropOnClose: false,
		clearOnOpen: !keepData,
	}

	s, err := NewBaseStorage(path, &cfg, cmdVisit)
	if err != nil {
		return nil, err
	}

	return &stgVisit{
		s: s,
	}, nil
}

// ------------------------------------------------------------------------

// Close closes the SQLite3 visit storage.
func (s *stgVisit) Close() error {
	return s.s.Close()
}

// ------------------------------------------------------------------------

// Clear removes all entries from the SQLite3 visit storage.
func (s *stgVisit) Clear() error {
	return s.s.Clear()
}

// ------------------------------------------------------------------------

// Len returns the number of visited requests in the SQLite3 visit storage.
func (s *stgVisit) Len() (uint, error) {
	return s.s.Len()
}

// ------------------------------------------------------------------------

// SetVisited stores a request ID that is visited by the Collector.
func (s *stgVisit) SetVisited(requestID uint64) error {
	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	_, err := s.s.stmts["insert"].Exec(requestID)

	return err
}

// ------------------------------------------------------------------------

// IsVisited returns true if the request was visited before.
func (s *stgVisit) IsVisited(requestID uint64) (bool, error) {
	var check int

	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	if err := s.s.stmts["select"].QueryRow(requestID).Scan(&check); err != nil {
		return false, err
	}

	return check == 1, nil
}

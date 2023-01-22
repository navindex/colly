package sqlite3

// ------------------------------------------------------------------------

type stgVisit struct {
	s *stgBase
}

// ------------------------------------------------------------------------

const defaultVisitTableName = "visits"

// ------------------------------------------------------------------------

var (
	cmdVisit = map[string]string{
		"create": `CREATE TABLE IF NOT EXISTS "<table>" ("key" TEXT PRIMARY KEY NOT NULL, "visits" INT)`,
		"drop":   `DROP TABLE IF EXISTS "<table>"`,
		"trim":   `DELETE FROM "<table>"`,
		"insert": `INSERT INTO "<table>" ("key", "visits") VALUES (?, 1) ON CONFLICT("key") DO UPDATE SET "visits" = "visits" + 1`,
		"select": `SELECT COALESCE("visits", 0) AS "visits" FROM "<table>" WHERE "key" = ?`,
		"delete": `DELETE FROM "<table>" WHERE "key" = ?`,
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

// AddVisit stores a request ID that is visited by the Collector.
func (s *stgVisit) AddVisit(key string) error {
	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	_, err := s.s.stmts["insert"].Exec(key)

	return err
}

// ------------------------------------------------------------------------

// PastVisits returns how many times the URL was visited before.
func (s *stgVisit) PastVisits(key string) (uint, error) {
	var visits int

	s.s.lock.Lock()
	err := s.s.stmts["select"].QueryRow(key).Scan(&visits)
	s.s.lock.Unlock()

	if err != nil {
		visits = 0
	}

	return uint(visits), err
}

// ------------------------------------------------------------------------

// Remove deletes a stored item by key.
func (s *stgVisit) Remove(key string) error {
	s.s.lock.Lock()
	_, err := s.s.stmts["delete"].Exec(key)
	s.s.lock.Unlock()

	return err
}

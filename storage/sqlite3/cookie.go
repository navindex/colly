package sqlite3

// ------------------------------------------------------------------------

type Storable interface {
	BinaryEncode() ([]byte, error)
}

type stgCookie struct {
	s *stgBase
}

// ------------------------------------------------------------------------

const defaultCookieJarName = "cookie_jar"

// ------------------------------------------------------------------------

var (
	cmdCookie = map[string]string{
		"create": `CREATE TABLE IF NOT EXISTS "<table>" ("host" TEXT PRIMARY KEY, "cookies" BLOB) WITHOUT ROWID`,
		"drop":   `DROP TABLE IF EXISTS "<table>"`,
		"trim":   `DELETE FROM "<table>"`,
		"insert": `INSERT INTO "<table>" ("host", "cookies") VALUES (?, ?) ON CONFLICT("host") DO UPDATE SET "cookies" = "excluded"."cookies"`,
		"select": `SELECT "cookies" FROM "<table>" WHERE "host" = ?`,
		"delete": `DELETE FROM "<table>" WHERE "host" = ?`,
		"count":  `SELECT COUNT(*) FROM "<table>"`,
	}
)

// ------------------------------------------------------------------------

// NewCookieStorage returns a pointer to a newly created SQLite3 cookie storage.
func NewCookieStorage(path string, table string, keepData bool) (*stgCookie, error) {
	cfg := config{
		table:       setTable(table, defaultCookieJarName),
		dropOnClose: false,
		clearOnOpen: !keepData,
	}

	s, err := NewBaseStorage(path, &cfg, cmdCookie)
	if err != nil {
		return nil, err
	}

	return &stgCookie{
		s: s,
	}, nil
}

// ------------------------------------------------------------------------

// Close closes the SQLite3 cookie storage.
func (s *stgCookie) Close() error {
	return s.s.Close()
}

// ------------------------------------------------------------------------

// Clear removes all entries from the SQLite3 cookie storage.
func (s *stgCookie) Clear() error {
	return s.s.Clear()
}

// ------------------------------------------------------------------------

// Len returns the number of hosts in the SQLite3 cookie storage.
func (s *stgCookie) Len() (uint, error) {
	return s.s.Len()
}

// ------------------------------------------------------------------------

// Set stores cookies for a given host.
func (s *stgCookie) Set(key string, data []byte) error {
	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	_, err := s.s.stmts["insert"].Exec(key, data)

	return err
}

// ------------------------------------------------------------------------

// Get retrieves stored cookies for a given host.
func (s *stgCookie) Get(key string) ([]byte, error) {
	var data = []byte{}

	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	err := s.s.stmts["select"].QueryRow(key).Scan(&data)

	return data, err
}

// ------------------------------------------------------------------------

// Remove deletes stored cookies for a given host.
func (s *stgCookie) Remove(key string) error {
	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	_, err := s.s.stmts["delete"].Exec(key)

	return err
}

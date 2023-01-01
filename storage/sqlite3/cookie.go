package sqlite3

import (
	"net/http"
	"net/url"

	"github.com/gocolly/colly/v2/storage"
)

// ------------------------------------------------------------------------

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

// SetCookies stores cookies for a given host.
func (s *stgCookie) SetCookies(u *url.URL, cookies []*http.Cookie) error {
	data, err := storage.CookiesToBytes(cookies)
	if err != nil {
		return err
	}

	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	_, err = s.s.stmts["insert"].Exec(u.Host, data)

	return err
}

// ------------------------------------------------------------------------

// Cookies retrieves stored cookies for a given host.
func (s *stgCookie) Cookies(u *url.URL) ([]*http.Cookie, error) {
	var data = []byte{}

	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	if err := s.s.stmts["select"].QueryRow(u.Host).Scan(&data); err != nil {
		return nil, err
	}

	return storage.BytesToCookies(data)
}

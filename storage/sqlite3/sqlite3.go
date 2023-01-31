package sqlite3

import (
	"colly/storage"
	"database/sql"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// ------------------------------------------------------------------------

// dbconn encapsulates the SQLite3 database handle
type dbconn struct {
	path     string
	dbh      *sql.DB // Database handle
	useCount uint16
}

// stgBase is a generic SQLite3 storage
type stgBase struct {
	db     *dbconn
	stmts  map[string]*sql.Stmt
	config *config
	lock   *sync.Mutex
	closed bool
}

// Storage config
type config struct {
	table       string
	dropOnClose bool
	clearOnOpen bool
}

// ------------------------------------------------------------------------

const placeholderTable = "<table>"

// ------------------------------------------------------------------------

// Database list indexed by file path
var connections = map[string]*dbconn{}

// Maximum number of storages connected to the same database.
// 0 value means no limit.
var maxUseCount uint16 = 100

var connLock = sync.Mutex{}

// ------------------------------------------------------------------------

// connect attaches a storage to a database
func connect(path string) (*dbconn, error) {
	if path == "" {
		return nil, storage.ErrBlankPath
	}

	connLock.Lock()
	defer connLock.Unlock()

	conn, present := connections[path]
	if !present {
		dbh, err := sql.Open("sqlite3", path)
		if err != nil {
			return nil, err
		}

		if err = dbh.Ping(); err != nil {
			dbh.Close()

			return nil, err
		}

		conn = &dbconn{
			path:     path,
			dbh:      dbh,
			useCount: 0,
		}
		connections[path] = conn
	}

	if maxUseCount > 0 && conn.useCount >= maxUseCount {
		return nil, storage.ErrStorageLimit
	}
	conn.useCount++

	return conn, nil
}

// ------------------------------------------------------------------------

// disconnect detaches a storage from the database
// and closes the database if no more storages are connected
func (dbc *dbconn) disconnect() {
	connLock.Lock()
	defer connLock.Unlock()

	dbc.useCount--

	// Remove dbc if this was the last connected storage
	if dbc.useCount <= 0 {
		dbc.dbh.Close()
		delete(connections, dbc.path)
	}
}

// ------------------------------------------------------------------------

// NewBaseStorage returns a pointer to a newly created SQLite3 base storage.
func NewBaseStorage(path string, config *config, commands map[string]string) (*stgBase, error) {
	if config == nil || commands == nil {
		return nil, storage.ErrMissingParams
	}

	if config.table == "" {
		return nil, storage.ErrBlankTableName
	}

	db, err := connect(path)
	if err != nil {
		return nil, err
	}

	s := &stgBase{
		db:     db,
		config: config,
		lock:   &sync.Mutex{},
		closed: false,
	}

	if err := s.addStatements(commands); err != nil {
		s.db.disconnect()

		return nil, err
	}

	// Create the tables if necessary
	if err := s.Cmd("create"); err != nil {
		s.db.disconnect()

		return nil, err
	}

	// Clear the data if required
	if s.config.clearOnOpen {
		if err := s.Cmd("trim"); err != nil {
			s.db.disconnect()

			return nil, err
		}
	}

	return s, nil
}

// ------------------------------------------------------------------------

// Close closes the SQLite3 storage.
func (s *stgBase) Close() error {
	var err error

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.config.dropOnClose {
		err = s.Cmd("drop")
	}

	s.db.disconnect()
	s.db = nil
	s.closed = true

	return err
}

// ------------------------------------------------------------------------

// Clear removes all entries from the SQLite3 storage.
func (s *stgBase) Clear() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.Cmd("trim")
}

// ------------------------------------------------------------------------

// Len returns the number of entries in the SQLite3 storage.
func (s *stgBase) Len(args ...any) (uint, error) {
	cmd := "count"

	stmt, present := s.stmts[cmd]
	if !present {
		return 0, storage.ErrMissingCmd(cmd)
	}

	var count int

	s.lock.Lock()
	defer s.lock.Unlock()

	if err := stmt.QueryRow(args...).Scan(&count); err != nil {
		return 0, err
	}

	return uint(count), nil
}

// ------------------------------------------------------------------------

func (s *stgBase) Cmd(cmd string, args ...any) error {
	stmt, present := s.stmts[cmd]
	if !present {
		return storage.ErrMissingCmd(cmd)
	}

	_, err := stmt.Exec(args...)

	return err
}

// ------------------------------------------------------------------------

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// func (s *stgBase) Exec(query string, args ...any) (sql.Result, error) {
// 	return s.db.dbh.Exec(query, args...)
// }

// ------------------------------------------------------------------------

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// func (s *stgBase) QueryRow(query string, args ...any) *sql.Row {
// 	return s.db.dbh.QueryRow(query, args...)
// }

// ------------------------------------------------------------------------

func (s *stgBase) addStatements(commands map[string]string) error {
	s.stmts = map[string]*sql.Stmt{}

	if len(commands) == 0 {
		return storage.ErrMissingStatement
	}

	for key, cmd := range commands {
		stmt, err := s.db.dbh.Prepare(strings.ReplaceAll(cmd, placeholderTable, s.config.table))
		if err != nil {
			return err
		}
		s.stmts[key] = stmt
	}

	return nil
}

// ------------------------------------------------------------------------

func setTable(table string, fallback string) string {
	table = strings.TrimSpace(table)

	if table == "" {
		table = strings.TrimSpace(fallback)
	}

	return strings.ReplaceAll(table, " ", "_")
}

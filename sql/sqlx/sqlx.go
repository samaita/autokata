package sqlx

import "github.com/jmoiron/sqlx"

func (s *db) Connect(driverName, dataSourceName string) (*sqlx.DB, error) {
	return sqlx.Connect(driverName, dataSourceName)
}

func (s *db) Beginx() (*sqlx.Tx, error) {
	return s.Conn.Beginx()
}

func (s *db) Get(dest interface{}, query string, args ...interface{}) error {
	return s.Conn.Get(dest, query, args...)
}

func (s *db) Select(dest interface{}, query string, args ...interface{}) error {
	return s.Conn.Select(dest, query, args...)
}

func (s *db) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return s.Conn.Queryx(query, args...)
}

func (s *db) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return s.Conn.QueryRowx(query, args...)
}

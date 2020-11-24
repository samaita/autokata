package sqlx

import "github.com/jmoiron/sqlx"

type DB interface {
	Beginx() (*sqlx.Tx, error)
	Connect(string, string) (*sqlx.DB, error)
	Get(interface{}, string, ...interface{}) error
	Queryx(string, ...interface{}) (*sqlx.Rows, error)
	QueryRowx(string, ...interface{}) *sqlx.Row
	Select(interface{}, string, ...interface{}) error
}

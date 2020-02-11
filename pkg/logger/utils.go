package customloggerpg

import (
	"database/sql"
	"fmt"
)

//OpenDB ...
func OpenDB(host string, port int, dbname string) (*sql.DB, func() error, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable", host, port, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, nil, err
	}

	return db, db.Close, err
}

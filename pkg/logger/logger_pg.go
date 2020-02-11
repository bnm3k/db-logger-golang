package customloggerpg

import (
	"database/sql"
	"log"
	"regexp"
)

type customOutPG struct {
	db         *sql.DB
	insertStmt string
	lp         *logParser
}

func (c *customOutPG) Write(log []byte) (int, error) {
	pl, err := c.lp.parseLog(string(log))
	if err != nil {
		return -1, err
	}
	_, err = c.db.Exec(c.insertStmt, pl.Prefix, pl.LogTime, pl.File, pl.Payload)

	return len(log), err
}

func newcustomOutPG(db *sql.DB) *customOutPG {
	return &customOutPG{
		db:         db,
		insertStmt: "insert into log(prefix, log_time, file, payload) values ($1, $2, $3, $4)",
		lp:         newLogParser(),
	}
}

//NewCustomLoggerPG ...
func NewCustomLoggerPG(prefix string, flag int, db *sql.DB) (*log.Logger, error) {
	//ensure prefix is of solely alphanumeric characters
	match, err := regexp.MatchString("^\\w+$", prefix)
	if err != nil || match == false {
		return nil, ErrInvalidPrefix
	}
	cOut := newcustomOutPG(db)
	return log.New(cOut, prefix+"\t", flag), nil

}

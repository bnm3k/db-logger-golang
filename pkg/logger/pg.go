package customlogger

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

//ErrInvalidPrefix ...
var ErrInvalidPrefix = errors.New("Invalid Prefix")

//ErrInvalidLog ..
var ErrInvalidLog = errors.New("Invalid Log. Unable to Parse")

type customOut struct {
	r          *regexp.Regexp
	db         *sql.DB
	insertStmt string
}

type parsedLog struct {
	Prefix  string
	LogTime time.Time
	File    string
	Payload string
}

func (c *customOut) Write(log []byte) (n int, err error) {
	pl, err := c.parseLog(string(log))
	if err != nil {
		fmt.Println(err)
	} else {
		_, err = c.db.Exec(c.insertStmt, pl.Prefix, pl.LogTime, pl.File, pl.Payload)
		if err != nil {
			// TODO handle err better
			fmt.Println(err)
		}
	}
	return len(log), err
}

func newCustomOut(db *sql.DB) *customOut {
	return &customOut{
		r:          regexp.MustCompile(`^(\w+)\s+(\d{4}\/\d{2}\/\d{2}\s)?(\d{2}:\d{2}:\d{2}(\.\d+)?\s)?(.*\.go:\d+:\s)?([\w\n]*)`),
		db:         db,
		insertStmt: "insert into log(prefix, log_time, file, payload) values ($1, $2, $3, $4)",
	}
}

func (c *customOut) parseLog(str string) (parsedLog, error) {
	var err error = nil
	var pl parsedLog

	matches := c.r.FindStringSubmatch(str)
	if matches != nil {
		var logTime time.Time
		logTime, err = parseLogTime(matches[2], matches[3])
		pl = parsedLog{
			Prefix:  matches[1],
			LogTime: logTime,
			File:    strings.TrimSpace(matches[5]),
			Payload: matches[6],
		}
	} else {
		err = ErrInvalidLog
	}

	if err != nil {
		return parsedLog{}, ErrInvalidLog
	}
	return pl, nil
}

func parseLogTime(dateVal, timeVal string) (time.Time, error) {
	now := time.Now()
	var t time.Time
	var err error = nil
	if dateVal == "" && timeVal == "" {
		// No date val. No time val
		return now, nil
	} else if dateVal == "" {
		// Only time val provided"
		y, m, d := now.Date()
		dtValStr := fmt.Sprintf("%v/%02d/%02d %s", y, m, d, timeVal)
		t, err = time.Parse("2006/01/02 15:04:05.999999 ", dtValStr)
	} else if timeVal == "" {
		// Only date val provided"
		t, err = time.Parse("2006/01/02 ", dateVal)
	} else {
		// Both date val and time val provided
		dtValStr := fmt.Sprintf("%s%s", dateVal, timeVal)
		t, err = time.Parse("2006/01/02 15:04:05.999999 ", dtValStr)
	}

	return t, err
}

//NewCustomLoggerPG ...
func NewCustomLoggerPG(prefix string, flag int, db *sql.DB) (*log.Logger, error) {
	//ensure prefix is of solely alphanumeric characters
	match, err := regexp.MatchString("^\\w+$", prefix)
	if err != nil || match == false {
		return nil, ErrInvalidPrefix
	}
	cOut := newCustomOut(db)
	return log.New(cOut, prefix+"\t", flag), nil

}

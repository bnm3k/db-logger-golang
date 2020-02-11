package customloggerpg

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

//ErrInvalidPrefix ...
var ErrInvalidPrefix = errors.New("Invalid Prefix")

//ErrInvalidLog ..
var ErrInvalidLog = errors.New("Invalid Log. Unable to Parse")

type logParser struct {
	logRegexMatch *regexp.Regexp
}

type parsedLog struct {
	Prefix  string    `json:"prefix"`
	LogTime time.Time `json:"timestamp"`
	File    string    `json:"file"`
	Payload string    `json:"payload"`
}

func (pl *parsedLog) String() string {
	s := fmt.Sprintf("%s::[%s, %s]\n\t%s", pl.Prefix, pl.LogTime, pl.File, pl.Payload)
	return s
}

func (lp *logParser) parseLog(str string) (*parsedLog, error) {
	var err error = nil
	var pl *parsedLog

	matches := lp.logRegexMatch.FindStringSubmatch(str)
	if matches != nil {
		var logTime time.Time
		logTime, err = parseLogTime(matches[2], matches[3])
		pl = &parsedLog{
			Prefix:  matches[1],
			LogTime: logTime,
			File:    strings.TrimSpace(matches[5]),
			Payload: matches[6],
		}
	} else {
		err = ErrInvalidLog
	}

	if err != nil {
		return nil, ErrInvalidLog
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

func newLogParser() *logParser {
	return &logParser{
		logRegexMatch: regexp.MustCompile(`^(\w+)\s+(\d{4}\/\d{2}\/\d{2}\s)?(\d{2}:\d{2}:\d{2}(\.\d+)?\s)?(.*\.go:\d+:\s)?([\w\W]*)`),
	}
}

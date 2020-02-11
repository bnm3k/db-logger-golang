package customloggerpg

import (
	"database/sql"
	"fmt"
)

//LogDAO ...
type LogDAO struct {
	db *sql.DB
}

//NewLogDAO ...
func NewLogDAO(db *sql.DB) LogDAO {
	return LogDAO{db: db}
}

func (l *LogDAO) latestHelper(stmt, prefix string) ([]string, error) {
	var err error
	var rows *sql.Rows
	if prefix != "" {
		rows, err = l.db.Query(stmt, prefix)
	} else {
		rows, err = l.db.Query(stmt)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []string
	for rows.Next() {
		log := &parsedLog{}
		err = rows.Scan(&log.Prefix, &log.LogTime, &log.File, &log.Payload)
		if err != nil {
			return nil, err
		}
		logs = append(logs, fmt.Sprintf("%v", log))
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

//Latest1Day ...
func (l *LogDAO) Latest1Day() ([]string, error) {
	stmt := `select prefix, log_time, file, payload from log where log_time >= now() - '1 day'::interval`
	return l.latestHelper(stmt, "")
}

//Latest1Week ...
func (l *LogDAO) Latest1Week() ([]string, error) {
	stmt := `select prefix, log_time, file, payload from log where log_time >= now() - '1 week'::interval`
	return l.latestHelper(stmt, "")
}

//Latest1DayWithPrefix ...
func (l *LogDAO) Latest1DayWithPrefix(prefix string) ([]string, error) {
	stmt := `
		select prefix, log_time, file, payload
		from log
		where log_time >= now() - '24 hours'::interval and prefix = $1`
	return l.latestHelper(stmt, prefix)
}

//Latest1WeekWithPrefix ...
func (l *LogDAO) Latest1WeekWithPrefix(prefix string) ([]string, error) {
	stmt := `
		select prefix, log_time, file, payload
		from log
		where log_time >= now() - '1 Week'::interval and prefix = $1`
	return l.latestHelper(stmt, prefix)
}

//ClearLogs ...
func (l *LogDAO) ClearLogs() error {
	_, err := l.db.Exec("truncate log")
	return err
}

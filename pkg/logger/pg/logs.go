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

func (l *LogDAO) latestHelper(stmt string) ([]string, error) {
	rows, err := l.db.Query(stmt)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var logs []string
	for rows.Next() {
		l := &parsedLog{}
		err = rows.Scan(&l.Prefix, &l.LogTime, &l.File, &l.Payload)
		if err != nil {
			return nil, err
		}
		logs = append(logs, fmt.Sprintf("%v", l))
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

//Latest1Day ...
func (l *LogDAO) Latest1Day() ([]string, error) {
	stmt := `select prefix, log_time, file, payload from log where log_time >= now() - '1 day'::interval`
	return l.latestHelper(stmt)
}

//Latest1Week ...
func (l *LogDAO) Latest1Week() ([]string, error) {
	stmt := `select prefix, log_time, file, payload from log where log_time >= now() - '1 week'::interval`
	return l.latestHelper(stmt)
}

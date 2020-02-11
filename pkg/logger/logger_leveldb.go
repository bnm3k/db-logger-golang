package customloggerpg

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/syndtr/goleveldb/leveldb"
)

type customOutLevelDB struct {
	db *leveldb.DB
	lp *logParser
}

func newCustomOutLevelDB(leveldbPath string) (*customOutLevelDB, error) {
	db, err := leveldb.OpenFile(leveldbPath, nil)
	if err != nil {
		return nil, err
	}
	return &customOutLevelDB{
		db: db,
		lp: newLogParser(),
	}, nil
}

func (c *customOutLevelDB) Write(log []byte) (int, error) {
	pl, err := c.lp.parseLog(string(log))
	if err != nil {
		return -1, err
	}
	key := fmt.Sprintf("%s!%020d", pl.Prefix, pl.LogTime.Unix())
	plJSON, err := json.Marshal(pl)
	if err != nil {
		return -1, err
	}
	err = c.db.Put([]byte(key), plJSON, nil)

	return len(log), err
}

//NewCustomLoggerLevelDB ...
func NewCustomLoggerLevelDB(prefix string, flag int, leveldbPath string) (*log.Logger, error) {
	//ensure prefix is of solely alphanumeric characters
	match, err := regexp.MatchString("^\\w+$", prefix)
	if err != nil || match == false {
		return nil, ErrInvalidPrefix
	}
	cOutLevelDB, err := newCustomOutLevelDB(leveldbPath)
	if err != nil {
		return nil, err
	}
	return log.New(cOutLevelDB, prefix+"\t", flag), nil
}

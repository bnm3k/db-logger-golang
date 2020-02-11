package customloggerpg

import (
	"database/sql"
	"io"
	"log"
	"regexp"
	"sync"
)

type customOutConc struct {
	logsCh chan []byte
	wg     *sync.WaitGroup
}

func (cc *customOutConc) Write(log []byte) (int, error) {
	cc.logsCh <- log
	return 0, nil
}

func (cc *customOutConc) Close() {
	close(cc.logsCh)
	cc.wg.Wait()
}

func logWorker(out io.Writer, logsCh <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for log := range logsCh {
		out.Write(log)
	}
}

func newOutWrapperConc(out io.Writer, bufSize, logWorkers int) *customOutConc {
	logsCh := make(chan []byte, bufSize)

	var wg sync.WaitGroup
	for i := 0; i < logWorkers; i++ {
		wg.Add(1)
		go logWorker(out, logsCh, &wg)
	}

	return &customOutConc{
		logsCh: logsCh,
		wg:     &wg,
	}
}

func newCustomLoggerConcHelper(prefix string, flag int, cout io.Writer) (*log.Logger, func(), error) {
	//ensure prefix is of solely alphanumeric characters
	match, err := regexp.MatchString("^\\w+$", prefix)
	if err != nil || match == false {
		return nil, nil, ErrInvalidPrefix
	}
	bufSize := 50
	logWorkers := 10
	wrappedCout := newOutWrapperConc(cout, bufSize, logWorkers)
	return log.New(wrappedCout, prefix+"\t", flag), wrappedCout.Close, nil

}

//NewCustomLoggerPGConc ...
func NewCustomLoggerPGConc(prefix string, flag int, db *sql.DB) (*log.Logger, func(), error) {
	pgcout := newCustomOut(db)
	return newCustomLoggerConcHelper(prefix, flag, pgcout)
}

//NewCustomLoggerLevelDBConc ...
func NewCustomLoggerLevelDBConc(prefix string, flag int, leveldbPath string) (*log.Logger, func(), error) {
	leveldbcout, err := newCustomOutLevelDB(leveldbPath)
	if err != nil {
		return nil, nil, err
	}
	return newCustomLoggerConcHelper(prefix, flag, leveldbcout)
}

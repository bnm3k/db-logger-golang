package customloggerpg

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	_ "github.com/lib/pq"
)

const sampleText = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func doSomeLogging(logger *log.Logger, wg *sync.WaitGroup, n int) {
	defer wg.Done()
	for i := 0; i < n; i++ {
		logger.Println(sampleText)
	}
}

func Benchmark_NewCustomLoggerPG(b *testing.B) {
	db, closeDB, err := OpenDB("localhost", 5432, "logging_golang")
	checkErr(err)
	defer closeDB()

	pglogs := NewLogDAO(db)
	infoLog, err := NewCustomLoggerPG("INFO", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile, db)
	checkErr(err)

	for n := 0; n < b.N; n++ {
		infoLog.Println(sampleText)
	}

	pglogs.ClearLogs()
}

func Benchmark_NewCustomLoggerPGConc(b *testing.B) {
	db, closeDB, err := OpenDB("localhost", 5432, "logging_golang")
	checkErr(err)
	defer closeDB()

	pglogs := NewLogDAO(db)
	infoLog, flushFn, err := NewCustomLoggerPGConc("INFO", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile, db)
	checkErr(err)
	defer flushFn()

	for n := 0; n < b.N; n++ {
		infoLog.Println(sampleText)
	}

	pglogs.ClearLogs()
}

func Benchmark_NewCustomLoggerPG_10_goroutines(b *testing.B) {
	db, closeDB, err := OpenDB("localhost", 5432, "logging_golang")
	checkErr(err)
	defer closeDB()

	pglogs := NewLogDAO(db)
	infoLog, err := NewCustomLoggerPG("INFO", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile, db)
	checkErr(err)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go doSomeLogging(infoLog, &wg, b.N)
	}
	wg.Wait()

	pglogs.ClearLogs()
}

func Benchmark_NewCustomLoggerPGConc_10_goroutines(b *testing.B) {
	db, closeDB, err := OpenDB("localhost", 5432, "logging_golang")
	checkErr(err)
	defer closeDB()

	pglogs := NewLogDAO(db)
	infoLog, flushFn, err := NewCustomLoggerPGConc("INFO", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile, db)
	checkErr(err)
	defer flushFn()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go doSomeLogging(infoLog, &wg, b.N)
	}
	wg.Wait()

	pglogs.ClearLogs()
}

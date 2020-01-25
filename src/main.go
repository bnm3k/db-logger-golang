package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	logpg "github.com/nagamocha3000/db-logger-golang/pkg/logger/pg"
)

func openDB(host string, port int, dbname string) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable", host, port, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}

func main() {
	checkErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	db, err := openDB("localhost", 5432, "logging_golang")
	checkErr(err)

	infoLog, err := logpg.NewCustomLoggerPG("ERROR", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile, db)
	checkErr(err)

	infoLog.Println("hello world logging some error stuff")
	pglogs := logpg.NewLogDAO(db)
	logs, err := pglogs.Latest1DayWithPrefix("ERROR")
	for _, log := range logs {
		fmt.Println(log)
	}
}

package main

import (
	"os"

	_ "github.com/lib/pq"
	logcli "github.com/nagamocha3000/db-logger-golang/pkg/cli"
	logpg "github.com/nagamocha3000/db-logger-golang/pkg/logger"
)

func main() {

	//connect to db
	db, closeDB, err := logpg.OpenDB("localhost", 5432, "logging_golang")
	defer closeDB()
	checkErr(err)

	//get logs DAO
	pglogs := logpg.NewLogDAO(db)

	//setup CLI
	app := logcli.SetupCLI(pglogs)
	err = app.Run(os.Args)
	checkErr(err)

	/*
			//using custom logger
		errLog, err := logpg.NewCustomLoggerPG("ERROR", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile, db)
		checkErr(err)
		errLog.Println("hello world logging some error stuff")
	*/
}

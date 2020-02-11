package main

import (
	"log"
	"os"

	_ "github.com/lib/pq"
	logcli "github.com/nagamocha3000/db-logger-golang/pkg/cli"
	logpg "github.com/nagamocha3000/db-logger-golang/pkg/logger"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	//connect to db
	db, closeDB, err := logpg.OpenDB("localhost", 5432, "logging_golang")
	defer closeDB()
	checkErr(err)

	//setup CLI
	app := logcli.SetupCLI(db)
	err = app.Run(os.Args)
	checkErr(err)

}

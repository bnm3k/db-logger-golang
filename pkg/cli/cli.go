package clilogs

import (
	"fmt"
	"log"
	"sort"

	logpg "github.com/nagamocha3000/db-logger-golang/pkg/logger"
	"github.com/urfave/cli/v2"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//SetupCLI ...
func SetupCLI(pglogs logpg.LogDAO) *cli.App {
	var dbName string
	app := &cli.App{
		Name:    "logs_cli",
		Usage:   "Handle app logs stored in pg via cli",
		Version: "v1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "dbName",
				Aliases:     []string{"n"},
				Value:       "logging_golang",
				Usage:       "set `DBNAME`",
				Destination: &dbName,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "clear_logs",
				Aliases: []string{"c"},
				Usage:   "clear all logs by truncating log table",
				Action: func(c *cli.Context) error {
					fmt.Printf("clearing the logs %q\n", dbName)
					err := pglogs.ClearLogs()
					return err
				},
			},
			{
				Name:    "print_logs",
				Aliases: []string{"p"},
				Usage:   "print all logs from past timeframe",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "prefix", Aliases: []string{"p"}, Value: "ERROR"},
				},
				Action: func(c *cli.Context) error {
					logs, err := pglogs.Latest1DayWithPrefix(c.String("prefix"))
					if err != nil {
						return err
					}
					for _, log := range logs {
						fmt.Println(log)
					}
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Println("welcome to logging cli")
			return nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	return app
}

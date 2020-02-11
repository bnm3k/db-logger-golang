package clilogs

import (
	"database/sql"
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
func SetupCLI(db *sql.DB) *cli.App {
	pglogs := logpg.NewLogDAO(db)
	app := &cli.App{
		Name:    "logs_cli",
		Usage:   "Handle app logs stored in pg via cli",
		Version: "v1.0.0",
		Commands: []*cli.Command{
			{
				Name:    "clear_logs",
				Aliases: []string{"c"},
				Usage:   "clear all logs by truncating log table",
				Action: func(c *cli.Context) error {
					fmt.Println("clearing the logs from db")
					err := pglogs.ClearLogs()
					return err
				},
			},
			{
				Name:  "print_logs",
				Usage: "print all logs from past timeframe",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "prefix", Aliases: []string{"p"}, Value: "INFO"},
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
			{
				Name:    "add_log",
				Aliases: []string{"a"},
				Usage:   "add given log",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "prefix", Aliases: []string{"p"}, Value: "INFO"},
					&cli.StringFlag{Name: "log", Aliases: []string{"l"}, Value: ""},
				},
				Action: func(c *cli.Context) error {
					prefix := c.String("prefix")
					logStr := c.String("log")
					if logStr == "" {
						fmt.Println("provide log to add")
						return nil
					}
					flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile

					logger, flush, err := logpg.NewCustomLoggerPGConc(prefix, flags, db)
					defer flush()
					if err == nil {
						logger.Print(logStr)
					}
					return err
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

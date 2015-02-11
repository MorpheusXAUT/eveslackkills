package main

import (
	"log"
	"os"
	"runtime"
	"time"

	"github.com/morpheusxaut/eveslackkills/database"
	"github.com/morpheusxaut/eveslackkills/misc"
	"github.com/morpheusxaut/eveslackkills/parser"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	config, err := misc.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: [%v]", err)
		os.Exit(2)
	}

	misc.SetupLogger(config.DebugLevel)

	db, err := database.SetupDatabase(config)
	if err != nil {
		misc.Logger.Criticalf("Failed to set up database: [%v]", err)
		os.Exit(2)
	}

	err = db.Connect()
	if err != nil {
		misc.Logger.Criticalf("Failed to connect to database: [%v]", err)
		os.Exit(2)
	}

	parse, err := parser.SetupParser(config, db, time.Minute*5)
	if err != nil {
		misc.Logger.Criticalf("Failed to set up parser: [%v]", err)
		os.Exit(2)
	}

	parse.Start()
}

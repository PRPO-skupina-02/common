package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/PRPO-skupina-02/common/database"
	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/lib/pq"
)

func main() {
	err := execute()
	if err != nil {
		log.Fatal(err)
	}
}

func execute() error {
	if len(os.Args) != 2 {
		return errors.New("Usage: dumpfixtures <dir>")
	}

	dsn := database.GetProdDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	dumper, err := testfixtures.NewDumper(
		testfixtures.DumpDatabase(db),
		testfixtures.DumpDialect("postgres"),
		testfixtures.DumpDirectory(os.Args[1]),
	)
	if err != nil {
		return err
	}

	err = dumper.Dump()
	if err != nil {
		return err
	}

	return nil
}

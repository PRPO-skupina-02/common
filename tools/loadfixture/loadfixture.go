package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/Masterminds/sprig/v3"
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
		return errors.New("Usage: loadfixture <dir>")
	}

	dsn := database.GetProdDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Template(),
		testfixtures.TemplateFuncs(sprig.FuncMap()),
		testfixtures.Directory(os.Args[1]),
		testfixtures.DangerousSkipTestDatabaseCheck(),
	)
	if err != nil {
		return err
	}

	err = fixtures.Load()
	if err != nil {
		return err
	}

	return nil
}

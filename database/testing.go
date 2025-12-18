package database

import (
	"database/sql"
	"embed"
	"fmt"
	"testing"

	"github.com/Masterminds/sprig/v3"
	"github.com/PRPO-skupina-02/common/config"
	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func GetTestDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.GetEnv("POSTGRES_IP"),
		config.GetEnv("POSTGRES_USERNAME"),
		config.GetEnv("POSTGRES_PASSWORD"),
		config.GetEnv("POSTGRES_TEST_DATABASE_NAME"),
		config.GetEnv("POSTGRES_PORT"))
}

func OpenTest() (*gorm.DB, error) {
	return Open(GetTestDSN())
}

func RecreateTestDatabase(t *testing.T) {
	prodDsn := GetProdDSN()

	prodDb, err := sql.Open("postgres", prodDsn)
	require.NoError(t, err)

	_, err = prodDb.Exec("DROP DATABASE IF EXISTS " + config.GetEnv("POSTGRES_TEST_DATABASE_NAME") + " WITH (FORCE)")
	require.NoError(t, err)

	_, err = prodDb.Exec("CREATE DATABASE " + config.GetEnv("POSTGRES_TEST_DATABASE_NAME"))
	require.NoError(t, err)

	err = prodDb.Close()
	require.NoError(t, err)
}

func PrepareTestDatabase(t *testing.T, fixtureFS, migrationsFS embed.FS) (*gorm.DB, *testfixtures.Loader) {
	RecreateTestDatabase(t)

	db, err := OpenTest()
	require.NoError(t, err)

	err = Migrate(db, migrationsFS)
	require.NoError(t, err)

	instance, err := db.DB()
	require.NoError(t, err)

	fixtures, err := testfixtures.New(
		testfixtures.Template(),
		testfixtures.TemplateFuncs(sprig.FuncMap()),
		testfixtures.Database(instance),
		testfixtures.Dialect("postgres"),
		testfixtures.FS(fixtureFS),
		testfixtures.Directory("fixtures"),
	)
	require.NoError(t, err)

	return db, fixtures
}

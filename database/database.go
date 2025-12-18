package database

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/PRPO-skupina-02/common/config"
	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetProdDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.GetEnv("POSTGRES_IP"),
		config.GetEnv("POSTGRES_USERNAME"),
		config.GetEnv("POSTGRES_PASSWORD"),
		config.GetEnv("POSTGRES_DATABASE_NAME"),
		config.GetEnv("POSTGRES_PORT"))
}

func OpenAndMigrateProd(migrationsFS embed.FS) (*gorm.DB, error) {
	return Open(GetProdDSN())
}

func OpenAndMigrate(dsn string, migrationsFS embed.FS) (*gorm.DB, error) {
	db, err := Open(dsn)
	if err != nil {
		return nil, err
	}

	err = Migrate(db, migrationsFS)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func OpenProd() (*gorm.DB, error) {
	return Open(GetProdDSN())
}

func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, err
	}
	slog.Debug("Successfully connected to database", "dsn", dsn)
	return db, nil
}

func Migrate(db *gorm.DB, migrationsFS embed.FS) error {
	instance, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := migratepostgres.WithInstance(instance, &migratepostgres.Config{})
	if err != nil {
		return err
	}

	migrationsFSDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	migrations, err := migrate.NewWithInstance("migrations", migrationsFSDriver, "postgres", driver)
	if err != nil {
		return err
	}

	err = migrations.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	slog.Debug("Database migrated successfully")
	return nil
}

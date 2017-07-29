package server

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

func Startup() {
	loadConfig()
	setupDatabase()
	err := runMigrations()
	if err != nil {
		log.Printf("Failed to run migrations: %v", err)
	}
}

var DB *sql.DB

func setupDatabase() {
	if Config.Database.Sqlite != nil {
		db, err := setupDatabaseSqlite()
		if err != nil {
			log.Fatalf("Failed to setup sqlite: %v", err)
		}
		DB = db
	}

	if DB == nil {
		log.Fatalf("Failed to setup database")
	}

	err := DB.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
}

func parseMigrationVersion(name string) int {
	split := strings.SplitN(name, "_", 2)
	if len(split) != 2 {
		return 0
	}
	version, err := strconv.Atoi(split[0])
	if err != nil {
		return 0
	}
	return version
}

func applyMigration(data string, version int) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("Failed to begin transaction: %v", err)
	}

	_, err = tx.Exec(string(data))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Failed to execute migration: %v", err)
	}

	_, err = tx.Exec("UPDATE version SET version=?", version)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Failed to set version: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %v", err)
	}

	return nil
}

func runMigrations() error {
	assert(DB != nil, "Trying to run migrations on nil database")

	path := filepath.Join(Config.Folders.Data, "migration")
	migrations, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("Failed to list migrations: %v", err)
	}

	dbVersion := 0
	DB.QueryRow(`SELECT version from version`).Scan(&dbVersion)
	loadedDbVersion := dbVersion

	log.Printf("Database is at version: %d", dbVersion)

	prevVersion := 0
	for _, migration := range migrations {
		_, file := filepath.Split(migration.Name())
		version := parseMigrationVersion(file)
		if version <= 0 {
			log.Printf("Invalid migration filename: %s", file)
			continue
		}
		if version != prevVersion+1 {
			log.Fatal("Migration ordering is not stable!")
		}
		prevVersion = version

		if dbVersion >= version {
			continue
		}

		fullName := filepath.Join(path, migration.Name())
		data, err := ioutil.ReadFile(fullName)
		if err != nil {
			log.Printf("Failed to read migration file '%s': %v", file, err)
			break
		}

		log.Printf("Applying migration: %s", file)
		err = applyMigration(string(data), version)
		if err != nil {
			log.Printf("Failed to apply migration '%s': %v", file, err)
			break
		}

		dbVersion = version
	}

	if loadedDbVersion != dbVersion {
		log.Printf("Database is at version: %d", dbVersion)
	}

	return nil
}

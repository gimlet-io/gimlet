package store

import (
	"database/sql"
	"os"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/ddl"

	"github.com/russross/meddler"

	"github.com/sirupsen/logrus"

	// MySQL driver
	_ "github.com/go-sql-driver/mysql"
	// PostgreSQL driver
	_ "github.com/lib/pq"
	// Sqlite driver
	_ "github.com/mattn/go-sqlite3"
)

// Store is used to access data
// from the sql/database driver with a relational database backend.
type Store struct {
	*sql.DB

	driver string
	config string
}

// New creates a database connection for the given driver and datasource
// and returns a new Store.
func New(driver, config string) *Store {
	return &Store{
		DB:     open(driver, config),
		driver: driver,
		config: config,
	}
}

// From returns a Store using an existing database connection.
func From(db *sql.DB) *Store {
	return &Store{DB: db}
}

// open opens a new database connection with the specified
// driver and connection string and returns a store.
func open(driver, config string) *sql.DB {
	db, err := sql.Open(driver, config)
	if err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("database connection failed")
	}
	if driver == "mysql" {
		// per issue https://github.com/go-sql-driver/mysql/issues/257
		db.SetMaxIdleConns(0)
	}

	setupMeddler(driver)

	if err := pingDatabase(db); err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("database ping attempts failed")
	}

	if err := setupDatabase(driver, db); err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("migration failed")
	}
	return db
}

// NewTest creates a new database connection for testing purposes.
// The database driver and connection string are provided by
// environment variables, with fallback to in-memory sqlite.
func NewTest() *Store {
	var (
		driver = "sqlite3"
		config = "file::memory:?cache=shared"
	)
	if os.Getenv("DATABASE_DRIVER") != "" {
		driver = os.Getenv("DATABASE_DRIVER")
		config = os.Getenv("DATABASE_CONFIG")
	}
	store := &Store{
		DB:     open(driver, config),
		driver: driver,
		config: config,
	}

	// if not in-memory DB, recreate tables between tests
	if driver != "sqlite3" {
		store.Exec(`
drop table migrations;
drop table users;
drop table commits;
drop table key_values;
`)
		setupDatabase(driver, store.DB)
	}

	return store
}

// helper function to ping the database with backoff to ensure
// a connection can be established before we proceed with the
// database setup and migration.
func pingDatabase(db *sql.DB) (err error) {
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			return
		}
		logrus.Infof("database ping failed. retry in 1s")
		time.Sleep(time.Second)
	}
	return
}

// helper function to setup the databsae by performing
// automated database migration steps.
func setupDatabase(driver string, db *sql.DB) error {
	return ddl.Migrate(driver, db)
}

// helper function to setup the meddler default driver
// based on the selected driver name.
func setupMeddler(driver string) {
	switch driver {
	case "sqlite3":
		meddler.Default = meddler.SQLite
	case "mysql":
		meddler.Default = meddler.MySQL
	case "postgres":
		meddler.Default = meddler.PostgreSQL
	}
}

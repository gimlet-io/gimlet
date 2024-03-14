package store

import (
	"database/sql"
	"os"
	"time"

	"github.com/gimlet-io/gimlet/pkg/dashboard/store/ddl"
	genericStore "github.com/gimlet-io/gimlet/pkg/store"

	"github.com/russross/meddler"

	"github.com/sirupsen/logrus"

	// PostgreSQL driver
	_ "github.com/lib/pq"
	// Sqlite driver
	_ "modernc.org/sqlite"
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
func New(driver, config, encryptionKey, encryptionKeyNew string) *Store {
	return &Store{
		DB:     open(driver, config, encryptionKey, encryptionKeyNew),
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
func open(driver, config, encryptionKey, encryptionKeyNew string) *sql.DB {
	db, err := sql.Open(driver, config)
	if err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("database connection failed")
	}

	setupMeddler(driver, encryptionKey, encryptionKeyNew)

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
func NewTest(encryptionKey, encryptionKeyNew string) *Store {
	var (
		driver = "sqlite"
		config = ":memory:"
	)
	if os.Getenv("DATABASE_DRIVER") != "" {
		driver = os.Getenv("DATABASE_DRIVER")
		config = os.Getenv("DATABASE_CONFIG")
	}
	store := &Store{
		DB:     open(driver, config, encryptionKey, encryptionKeyNew),
		driver: driver,
		config: config,
	}

	// if not in-memory DB, recreate tables between tests
	if driver != "sqlite" {
		store.Exec(`
drop table migrations;
drop table users;
drop table commits;
drop table key_values;
drop table environments;
drop table events;
drop table gitops_commits;
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
func setupMeddler(driver, encryptionKey, encryptionKeyNew string) {
	switch driver {
	case "sqlite":
		meddler.Default = meddler.SQLite
	case "postgres":
		meddler.Default = meddler.PostgreSQL
	}

	meddler.Register("encrypted", genericStore.EncryptionMeddler{EnryptionKey: encryptionKey, EncryptionKeyNew: encryptionKeyNew})
}

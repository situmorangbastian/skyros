package mysql_test

import (
	"database/sql"
	"path"
	"runtime"
	"strings"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
)

// Migration is struct for migration purpose
type Migration struct {
	Migrate *migrate.Migrate
}

// Up is method for migrate up database
func (m *Migration) Up() ([]error, bool) {
	err := m.Migrate.Up()
	if err != nil {
		return []error{err}, false
	}

	return []error{}, true
}

// Down is method for migrate down database
func (m *Migration) Down() ([]error, bool) {
	err := m.Migrate.Down()
	if err != nil {
		return []error{err}, false
	}

	return []error{}, true
}

// RunMigration is function to run the database migration (up and down)
func RunMigration(dbURI string) (*Migration, error) {
	_, filename, _, _ := runtime.Caller(0)

	migrationPath := path.Join(path.Dir(filename), "migrations_test")

	var dataPath []string
	dataPath = append(dataPath, "file://")
	dataPath = append(dataPath, migrationPath)

	pathToMigrate := strings.Join(dataPath, "")

	db, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, err
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{DatabaseName: mysqlDatabase})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		pathToMigrate,
		mysqlDriver,
		driver,
	)

	return &Migration{Migrate: m}, err
}

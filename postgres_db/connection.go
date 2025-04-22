package postgres_db

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"time"
)

// New creates and returns a new database connection pool.
// It ensures the database exists, configures connection settings, and validates the connection.
// Parameters:
// - dbHost: The hostname of the database server.
// - dbDriver: The database driver (e.g., "postgres").
// - dbPort: The port number of the database server.
// - dbUser: The username for database authentication.
// - dbPass: The password for database authentication.
// - dbName: The name of the database to connect to.
// - connMaxLifetime: The maximum lifetime of a connection in minutes (0 for no limit).
// - maxIdleConnections: The maximum number of idle connections in the pool.
// - maxConnections: The maximum number of open connections to the database (0 for no limit).
// Returns:
// - *sql.DB: A pointer to the database connection pool.
// - error: An error if the connection setup fails.
func New(dbHost, dbDriver, dbPort, dbUser, dbPass, dbName string, connMaxLifetime, maxIdleConnections, maxConnections int) (*sql.DB, error) {

	if err := ensureDatabaseExists(dbHost, dbDriver, dbPort, dbUser, dbPass, dbName); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	if connMaxLifetime == 0 {
		db.SetConnMaxLifetime(time.Nanosecond)
	} else {
		db.SetConnMaxLifetime(time.Minute * time.Duration(connMaxLifetime))
	}

	db.SetMaxIdleConns(maxIdleConnections)

	if maxConnections != 0 {
		db.SetMaxOpenConns(maxConnections)
	}

	return db, nil
}

// ensureDatabaseExists checks if the specified database exists and creates it if it does not.
// Parameters:
// - dbHost: The hostname of the database server.
// - dbDriver: The database driver (e.g., "postgres").
// - dbPort: The port number of the database server.
// - dbUser: The username for database authentication.
// - dbPass: The password for database authentication.
// - dbName: The name of the database to check or create.
// Returns:
// - error: An error if the operation fails.
func ensureDatabaseExists(dbHost, dbDriver, dbPort, dbUser, dbPass, dbName string) error {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable", dbHost, dbPort, dbUser, dbPass)

	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	if err = db.QueryRow(query, dbName).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(dbName))
		if _, err = db.Exec(createDBQuery); err != nil {
			return err
		}
	}

	return nil
}

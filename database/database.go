package database

import (
	"context"
	"database/sql"
	"erp-excel/config"

	"fmt"
	"log"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

// Database interface
type Database interface {
	DB() *sql.DB
	ERPDatabase() *sql.DB // Add missing method declaration
	Close() error
	Ping() error
}

type database struct {
	db    *sql.DB
	erpDB *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config) (Database, error) {
	db, err := sql.Open("sqlserver", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	erpDB, err := sql.Open("sqlserver", cfg.GetERPDatabaseDSN())
	// log.Printf("Database connection: %v", erpDB)
	if err != nil {
		db.Close() // Close the first DB connection if ERP DB fails
		return nil, fmt.Errorf("error opening ERP database: %w", err)
	}

	// Kiểm tra kết nối ERP database
	if err := erpDB.Ping(); err != nil {
		db.Close() // Close both connections if ping fails
		erpDB.Close()
		return nil, fmt.Errorf("error pinging ERP database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close() // Close connections if ping fails
		erpDB.Close()
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return &database{
		db:    db,
		erpDB: erpDB,
	}, nil
}

// MustDatabase panics if database connection fails
func MustDatabase(cfg *config.Config) Database {
	db, err := NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Fatal database connection error: %s", err)
	}
	return db
}

// DB returns the database connection
func (d *database) DB() *sql.DB {
	return d.db
}

// ERPDatabase returns the ERP database connection
func (d *database) ERPDatabase() *sql.DB {
	return d.erpDB
}

// Close closes both database connections
func (d *database) Close() error {
	// Close both databases and track potential errors
	var errs []error

	if err := d.db.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing main database: %w", err))
	}

	if err := d.erpDB.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing ERP database: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing databases: %v", errs)
	}

	return nil
}

// Ping checks if the database connection is alive
func (d *database) Ping() error {
	// Ping both databases
	if err := d.db.Ping(); err != nil {
		return fmt.Errorf("error pinging main database: %w", err)
	}

	if err := d.erpDB.Ping(); err != nil {
		return fmt.Errorf("error pinging ERP database: %w", err)
	}

	return nil
}

// ExecuteScript runs a SQL script to initialize database
func (d *database) ExecuteScript(script string) error {
	_, err := d.db.Exec(script)
	if err != nil {
		return fmt.Errorf("error executing script: %w", err)
	}
	return nil
}

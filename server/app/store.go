package app

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

const NoRecords = "No records returned."

type Record struct {
	Key   string
	Value string
}

type Store struct {
	Filename   string
	Prefix     string
	connection *sql.DB
}

func (s *Store) TableName() string {
	return s.Prefix + "records"
}

func CreateStore(filename string) (*Store, error) {
	// Create Object
	s := &Store{
		Filename: filename,
	}

	// Get Connection
	err := s.GetConnection()
	if err != nil {
		return s, err
	}

	// Create Tables
	err = s.CreateTables()
	return s, err
}

// Load the Database Connection
func (s *Store) GetConnection() error {
	db, err := sql.Open("sqlite3", s.dbLocation())
	if err != nil {
		return err
	}

	s.connection = db
	return nil
}

func (s *Store) CreateTables() error {
	_, err := s.connection.Exec("CREATE TABLE IF NOT EXISTS " + s.TableName() + " (key TEXT, value TEXT)")
	return err
}

func (s *Store) dbLocation() string {
	return filepath.Join(os.Getenv("MLGBASE"), "data", s.Filename)
}

func (s *Store) Set(key string, value string) error {
	r, err := s.connection.Exec("UPDATE "+s.TableName()+" SET value = ? WHERE key = ?", value, key)
	if err != nil {
		return err
	}

	count, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		_, err := s.connection.Exec("INSERT INTO "+s.TableName()+" VALUES (?, ?)", key, value)
		return err
	}
	return nil
}

func (s *Store) GetDefault(key string, alt string) (string, error) {
	val, err := s.Get(key)
	if val == "" && err.Error() == NoRecords {
		return alt, nil
	}
	return val, err
}

func (s *Store) Get(key string) (string, error) {
	// Select Rows
	rows, err := s.connection.Query("SELECT value FROM "+s.TableName()+" WHERE key=?", key)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Iterate Through Rows
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return "", err
		}
		return value, nil
	}

	// Check for Errors
	if err := rows.Err(); err != nil {
		return "", err
	}
	return "", errors.New(NoRecords)
}

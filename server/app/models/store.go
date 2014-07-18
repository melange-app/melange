package models

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
	_, err := s.connection.Exec("CREATE TABLE IF NOT EXISTS " + s.TableName() + " (key TEXT, value BLOB)")
	return err
}

func (s *Store) dbLocation() string {
	return filepath.Join(os.Getenv("MLGBASE"), "data", s.Filename)
}

func (s *Store) SetBytes(key string, value []byte) error {
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

func (s *Store) Set(key string, value string) error {
	return s.SetBytes(key, []byte(value))
}

func (s *Store) GetDefault(key string, alt string) (string, error) {
	val, err := s.Get(key)
	if val == "" && err.Error() == NoRecords {
		return alt, nil
	}
	return val, err
}

func (s *Store) GetBytes(key string) ([]byte, error) {
	// Select Rows
	rows, err := s.connection.Query("SELECT value FROM "+s.TableName()+" WHERE key=?", key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate Through Rows
	for rows.Next() {
		var value []byte
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		return value, nil
	}

	// Check for Errors
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nil, errors.New(NoRecords)
}

func (s *Store) Get(key string) (string, error) {
	data, err := s.GetBytes(key)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

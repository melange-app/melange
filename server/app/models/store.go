package models

import (
	"github.com/huntaub/go-db"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

type Record struct {
	Key   string
	Value []byte
}

type Store struct {
	*sqlx.DB
	Filename string
	Prefix   string
	table    db.Table
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
	conn, err := sqlx.Open("sqlite3", s.dbLocation())
	if err != nil {
		return err
	}

	s.DB = conn
	return nil
}

func (s *Store) CreateTables() error {
	t, err := db.CreateTableFromStruct(s.TableName(), s, false, &Record{})
	if err == nil {
		s.table = t
	}
	return err
}

func (s *Store) dbLocation() string {
	return filepath.Join(os.Getenv("MLGBASE"), "data", s.Filename)
}

func (s *Store) SetBytes(key string, value []byte) error {
	r, err := s.NamedExec("UPDATE "+s.TableName()+" SET value = :value WHERE key = :key", map[string]interface{}{
		"value": value,
		"key":   key,
	})
	if err != nil {
		return err
	}

	count, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		_, err := s.table.Insert(&Record{
			Key:   key,
			Value: value,
		}).Exec(s)
		return err
	}
	return nil
}

func (s *Store) Set(key string, value string) error {
	return s.SetBytes(key, []byte(value))
}

func (s *Store) GetDefault(key string, alt string) (string, error) {
	val, err := s.Get(key)
	if val == "" && err != nil {
		return alt, err
	}
	return val, err
}

func (s *Store) GetBytes(key string) ([]byte, error) {
	// Select Rows
	result := &Record{}
	err := s.table.Get().Where("key", key).One(s, result)
	if err != nil {
		return nil, err
	}

	return result.Value, nil
}

func (s *Store) Get(key string) (string, error) {
	data, err := s.GetBytes(key)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

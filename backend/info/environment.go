// Package info provides information about the currently running
// Melange backend to the Controllers that need it.
package info

import (
	"path/filepath"

	"getmelange.com/backend/cache"
	"getmelange.com/backend/models"
	"getmelange.com/backend/models/db"
	"getmelange.com/backend/packaging"
)

// Environment represents information about the version of Melange
// that is currently running.
type Environment struct {
	Tables   *db.Tables
	Packager *packaging.Packager
	Manager  *cache.Manager

	cached bool

	Settings
}

func (e *Environment) createStore() (err error) {
	e.Tables, err = models.CreateTables(e.Store)
	return err
}

func (e *Environment) createPackager() error {
	packager := &packaging.Packager{
		API:    "http://www.getmelange.com/api",
		Plugin: filepath.Join(e.DataDirectory, "plugins"),
		Debug:  e.Debug,
	}

	return packager.CreatePluginDirectory()
}

func (e *Environment) createMessageManager() (err error) {
	e.Manager, err = cache.CreateManager(e.Tables, e.Store, e.Packager)
	return err
}

// Cache allows the environment to startup and load information that
// will be necessary for the functioning of the server.
func (e *Environment) Cache() error {
	if e.cached {
		panic("Should not be attempting to re-cache information in the environment.")
	}

	if err := e.createStore(); err != nil {
		return err
	}

	if err := e.createPackager(); err != nil {
		return err
	}

	if err := e.createMessageManager(); err != nil {
		return err
	}

	e.cached = true
	return nil
}

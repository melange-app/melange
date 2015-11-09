package models

import gdb "github.com/huntaub/go-db"

// DeveloperPlugin is analogous to the message that publishes a plugin to the world
//
// `plugins/chat`
type DeveloperPlugin struct {
	// Plugin Information
	Id          gdb.PrimaryKey `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Image       string         `json:image"`

	// Source Information
	Source string `json:"source"`
}

// DeveloperPluginVersion has version-specific information.
//
// `plugins/chat/0.0.1`
type DeveloperPluginVersion struct {
	Version   string `json:"version"`
	Changelog string `json:"changelog"`

	// Permissions Information
	Permissions     []DeveloperPluginPermissions `db:"-" json:"permissions"`
	PermissionsData []byte                       `json:"-"`

	// Extras Information
	Viewers     []DeveloperPluginViewer `db:"-" json:"viewers"`
	Tiles       []DeveloperPluginTile   `db:"-" json:"tiles"`
	ViewersData []byte                  `json:"-"`
	TilesData   []byte                  `json:"-"`

	// Data Information
	Data string `json:"data"`
}

type DeveloperPluginPermissions struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DeveloperPluginViewer struct {
	Id     string   `json:"id"`
	Name   string   `json:"name"`
	Hidden bool     `json:"hidden"`
	Type   []string `json:"type"`
	View   string   `json:"view"`
}

type DeveloperPluginTile struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Size        string `json:"size"`
	View        string `json:"view"`
}

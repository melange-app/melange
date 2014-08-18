package framework

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Plugin struct {
	Id          string              `json:"id"`
	Name        string              `json:"name"`
	Version     string              `json:"version"`
	Description string              `json:"description"`
	Permissions map[string][]string `json:"permissions"`
	Author      Author              `json:"author"`
	Homepage    string              `json:"homepage"`
	HideSidebar bool                `json:"hideSidebar"`
	Tiles       map[string]Tile     `json:"tiles"`
	Viewers     map[string]Viewer   `json:"viewers"`
}

type Viewer struct {
	Type   []string `json:"type"`
	View   string   `json:"view"`
	Hidden bool     `json:"hidden"`
}

type Tile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	View        string `json:"view"`
	Size        string `json:"size"`
	Click       bool   `json:"click"`
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Homepage string `json:"homepage"`
}

func AllPlugins(pluginPath string) ([]Plugin, error) {
	files, err := ioutil.ReadDir(pluginPath)
	if err != nil {
		return nil, err
	}

	output := []Plugin{}
	for _, v := range files {
		if _, err := os.Stat(filepath.Join(pluginPath, v.Name(), "package.json")); err == nil && v.IsDir() {
			packageJSON, err := os.Open(filepath.Join(pluginPath, v.Name(), "package.json"))
			if err != nil {
				fmt.Println(err)
				continue
			}

			var plugin Plugin
			decoder := json.NewDecoder(packageJSON)
			err = decoder.Decode(&plugin)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if plugin.Id != v.Name() {
				continue
			}

			output = append(output, plugin)

			packageJSON.Close()
		}
	}

	return output, nil
}

func LoadPlugins(pluginPath string) View {
	output, err := AllPlugins(pluginPath)
	if err != nil {
		fmt.Println("Error getting all plugins", err)
		return Error500
	}

	bytes, err := json.Marshal(output)
	if err != nil {
		return Error500
	}
	return &RawView{bytes, "application/json"}
}

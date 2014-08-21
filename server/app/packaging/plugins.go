package packaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"getmelange.com/app/framework"
	"getmelange.com/updater"

	"github.com/google/go-github/github"
)

type SimplePlugin struct {
	Id          string
	Name        string
	Description string
	Username    string
	Repository  string
	Installed   bool
}

func (p *Packager) CreatePluginDirectory() error {
	return os.MkdirAll(p.Plugin, os.ModeDir|os.ModePerm)
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (p *Packager) DecodeStore() ([]*SimplePlugin, error) {
	resp, err := http.Get(p.API + "/applications")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	var obj []*SimplePlugin
	err = dec.Decode(&obj)
	if err != nil {
		return nil, err
	}

	for _, v := range obj {
		if b, _ := exists(filepath.Join(p.Plugin, v.Id)); b {
			v.Installed = true
		}
	}

	return obj, nil
}

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

func (p *Packager) InstallPlugin(repo string) error {
	repoComponents := strings.Split(repo, "/")
	if len(repoComponents) < 2 || !strings.HasPrefix(repo, "http://github.com/") {
		return errors.New("Not a valid github repo url.")
	}

	// Get the Path Components
	user := repoComponents[len(repoComponents)-2]
	name := repoComponents[len(repoComponents)-1]

	gh := github.NewClient(nil)

	// Get the Release Version
	tags, _, err := gh.Repositories.ListTags(user, name, nil)
	if err != nil {
		return err
	}

	var latest *github.RepositoryTag
	for _, v := range tags {
		if strings.HasPrefix(*v.Name, "v") {
			latest = &v
			break
		}
	}

	if latest == nil {
		return errors.New("Can't install a repository without a release.")
	}

	// Got here

	// Ensure that the Package ID is still correct
	packageJSON := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/package.json", user, name, *latest.Name)
	fmt.Println(packageJSON)
	res, err := http.Get(packageJSON)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	plugin := &Plugin{}
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(plugin)
	if err != nil {
		fmt.Println("Error decoding body")
		return err
	}

	expected := fmt.Sprintf("com.github.%s.%s", user, name)
	if plugin.Id != expected {
		return fmt.Errorf("Couldn't get plugin as ids didn't match. Expected %s, got %s", expected, plugin.Id)
	}

	// Download the Zipball
	file, err := ioutil.TempFile("", "melange_plugin_download_")
	if err != nil {
		return err
	}
	defer file.Close()

	zipball, err := http.Get(*latest.ZipballURL)
	if err != nil {
		return err
	}
	defer zipball.Body.Close()

	n, err := io.Copy(file, zipball.Body)
	if err != nil {
		return err
	}

	// Extract the Zipball into the Plugin Directory
	pluginDir := filepath.Join(p.Plugin, plugin.Id)

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	base, err := updater.ExtractZip(file, n, p.Plugin)
	if err != nil {
		return err
	}
	file.Close()

	err = os.Rename(filepath.Join(p.Plugin, base), pluginDir)
	if err != nil {
		return err
	}

	os.Remove(file.Name())

	return nil
}

func (p *Packager) UninstallPlugin(id string) error {
	return os.RemoveAll(filepath.Join(p.Plugin, id))
}

func (p *Packager) AllPlugins() ([]Plugin, error) {
	files, err := ioutil.ReadDir(p.Plugin)
	if err != nil {
		return nil, err
	}

	output := []Plugin{}
	for _, v := range files {
		if _, err := os.Stat(filepath.Join(p.Plugin, v.Name(), "package.json")); err == nil && v.IsDir() {
			packageJSON, err := os.Open(filepath.Join(p.Plugin, v.Name(), "package.json"))
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

func (p *Packager) LoadPlugins() framework.View {
	output, err := p.AllPlugins()
	if err != nil {
		fmt.Println("Error getting all plugins", err)
		return framework.Error500
	}

	bytes, err := json.Marshal(output)
	if err != nil {
		return framework.Error500
	}
	return &framework.RawView{bytes, "application/json"}
}

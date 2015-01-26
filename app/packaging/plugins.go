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
	"strconv"
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
	Id            string                   `json:"id"`
	Name          string                   `json:"name"`
	Version       string                   `json:"version"`
	Description   string                   `json:"description"`
	Permissions   map[string][]string      `json:"permissions"`
	Notifications map[string]*Notification `json:"notifications"`
	Author        Author                   `json:"author"`
	Homepage      string                   `json:"homepage"`
	HideSidebar   bool                     `json:"hideSidebar"`
	Tiles         map[string]Tile          `json:"tiles"`
	Viewers       map[string]Viewer        `json:"viewers"`
}

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
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

func (p *Packager) verifyPlugin(latest *github.RepositoryTag, user string, name string) (*Plugin, error) {
	// Ensure that the Package ID is still correct
	packageJSON := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/package.json", user, name, *latest.Name)
	fmt.Println(packageJSON)
	res, err := http.Get(packageJSON)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	plugin := &Plugin{}
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(plugin)
	if err != nil {
		fmt.Println("Error decoding body")
		return nil, err
	}

	// Check that plugin has correct id.
	expected := fmt.Sprintf("com.github.%s.%s", user, name)
	if plugin.Id != expected {
		return nil, fmt.Errorf("Couldn't get plugin as ids didn't match. Expected %s, got %s", expected, plugin.Id)
	}

	// Check that plugin has correct version.
	if plugin.Version != strings.TrimPrefix(*latest.Name, "v") {
		return nil, fmt.Errorf("Plugin doesn't have correct version (%s) in the manifest file (%s).", strings.TrimPrefix(*latest.Name, "v"), plugin.Version)
	}

	return plugin, nil
}

func (p *Packager) getLatestPluginVersion(gh *github.Client, user, name string) (*github.RepositoryTag, error) {
	// Get the Release Version
	tags, _, err := gh.Repositories.ListTags(user, name, nil)
	if err != nil {
		return nil, err
	}

	var latest *github.RepositoryTag
	for _, v := range tags {
		if strings.HasPrefix(*v.Name, "v") {
			latest = &v
			break
		}
	}

	if latest == nil {
		return nil, errors.New("Can't install a repository without a release.")
	}

	return latest, nil
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
	latest, err := p.getLatestPluginVersion(gh, user, name)
	if err != nil {
		return err
	}

	plugin, err := p.verifyPlugin(latest, user, name)
	if err != nil {
		return err
	}

	// Download the Zipball
	file, err := p.getTempFile()
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

type semVer string

func (c semVer) after(b semVer) bool {
	comps0 := strings.Split(string(c), ".")
	comps1 := strings.Split(string(b), ".")

	if len(comps0) != 3 || len(comps1) != 3 {
		// You didn't pass in a semantic version.
		fmt.Println("A SemVer is incorrect c:", c, "b:", b)
		return false
	}

	major0, err1 := strconv.Atoi(comps0[0])
	minor0, err2 := strconv.Atoi(comps0[1])
	build0, err3 := strconv.Atoi(comps0[2])

	if err1 != nil || err2 != nil || err3 != nil {
		// SemVer is improperly formatted
		fmt.Println("Couldn't convert SemVer 1", c)
		return false
	}

	major1, err1 := strconv.Atoi(comps1[0])
	minor1, err2 := strconv.Atoi(comps1[1])
	build1, err3 := strconv.Atoi(comps1[2])

	if err1 != nil || err2 != nil || err3 != nil {
		// SemVer is improperly formatted
		fmt.Println("Couldn't convert SemVer 2", b)
		return false
	}

	if major0 > major1 {
		return true
	}

	if major0 < major1 {
		return false
	}

	if minor0 > minor1 {
		return true
	}

	if minor0 < minor1 {
		return false
	}

	if build0 > build1 {
		return true
	}

	return false
}

type PluginUpdate struct {
	Id         string
	Version    semVer
	Changelog  string
	Repository string
}

func (p *Packager) CheckForPluginUpdates() ([]*PluginUpdate, error) {
	// Fetch the correct plugin
	plugins, err := p.AllPlugins()
	if err != nil {
		return nil, err
	}

	var updates []*PluginUpdate
	for _, v := range plugins {
		u, err := p.pluginUpdates(&v)

		if err != nil {
			fmt.Println("Error checking for updates on plugin", v.Id, err)
			continue
		}

		if u.Version.after(semVer(v.Version)) {
			updates = append(updates, u)
		}
	}

	return updates, nil
}

func (p *Packager) pluginUpdates(thePlugin *Plugin) (*PluginUpdate, error) {
	idComp := strings.Split(thePlugin.Id, ".")
	if len(idComp) != 4 || idComp[0] != "com" || idComp[1] != "github" {
		return nil, fmt.Errorf("Unable to get plugin with malformed id (%s).", idComp)
	}

	// Get the repository associated with the plugin
	user := idComp[2]
	name := idComp[3]
	repository := fmt.Sprintf("http://github.com/%s/%s", user, name)

	gh := github.NewClient(nil)
	latest, err := p.getLatestPluginVersion(gh, user, name)
	if err != nil {
		return nil, err
	}

	newestPlugin, err := p.verifyPlugin(latest, user, name)
	if err != nil {
		return nil, err
	}

	changelog := ""
	func() {
		if latest.Commit.Message != nil {
			changelog = *latest.Commit.Message
		} else {
			commit, _, err := gh.Repositories.GetCommit(user, name, *latest.Commit.SHA)

			if err != nil {
				fmt.Println("Couldn't get commit.", err)
				return
			}

			if commit.Message == nil {
				return
			}

			changelog = *commit.Message
		}
	}()

	return &PluginUpdate{
		Id:         newestPlugin.Id,
		Version:    semVer(newestPlugin.Version),
		Changelog:  changelog,
		Repository: repository,
	}, nil
}

func (p *Packager) ExecuteUpdate(update *PluginUpdate) error {
	err := p.UninstallPlugin(update.Id)
	if err != nil {
		return err
	}

	return p.InstallPlugin(update.Repository)
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
				fmt.Println(v.Name(), "package.json file open error:", err)
				continue
			}

			var plugin Plugin
			decoder := json.NewDecoder(packageJSON)
			err = decoder.Decode(&plugin)
			if err != nil {
				fmt.Println(v.Name(), "package.json parse error:", err)
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

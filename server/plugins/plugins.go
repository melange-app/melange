package plugins

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type Server struct {
	Suffix  string
	Common  string
	Plugins string
}

func (p *Server) CommonURL() string {
	return p.Common + p.Suffix
}

func (p *Server) PluginURL() string {
	return p.Plugins + p.Suffix
}

func (p *Server) Run(port int) error {
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: &Router{p},
	}
	return s.ListenAndServe()
}

type Router struct {
	p *Server
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// Different Behavior Based on Host

	url := strings.Split(req.Host, ".melange")
	if len(url) != 2 || !(strings.HasPrefix(url[1], ":") || url[1] == r.p.Suffix) {
		WriteView(&HTTPError{403, "Cannot access this service out of Melange."}, res)
		return
	}
	mode := url[0]

	if strings.HasSuffix(mode, "plugins") {
		// Serve Plugin Files
		pluginPath := strings.TrimSuffix(mode, ".plugins")
		view := FindFile(filepath.Join("plugins", pluginPath), req.URL.Path)
		WriteView(&CSPWrapper{
			CSP: fmt.Sprintf("default-src 'self';"+
				"img-src *;"+
				"font-src 'self' %[1]s;"+
				"script-src 'self' 'unsafe-eval' %[1]s;"+
				"style-src 'self' 'unsafe-inline' %[1]s", r.p.CommonURL()),
			View: view,
		}, res)
	} else if mode == "common" {
		// Serve Library Files
		if req.URL.Path == "/main/theme" {
			// Load the Main Theme Files
			WriteView(FindFile("lib", filepath.Join("bootswatch-yeti", "3.1.1.css")), res)
		} else {
			dirs := strings.Split(req.URL.Path, "/")
			// No More Panics
			if len(dirs) != 4 {
				WriteView(&HTTPError{404, "Couldn't find file."}, res)
				return
			}
			typ, lib, version := dirs[1], dirs[2], dirs[3]
			view := FindFile("lib", filepath.Join(filepath.FromSlash(lib), version+"."+filepath.FromSlash(typ)))
			WriteView(view, res)
		}
	} else if mode == "app" {
		// Serve Application Files
		if req.URL.Path == "/plugins.json" {
			// Serve the Plugins.JSON Document
			WriteView(LoadPlugins(filepath.Join(os.Getenv("MLGBASE"), "plugins")), res)
		} else {
			// Serve Regular Files
			view := FindFile("client", req.URL.Path)
			WriteView(&CSPWrapper{
				CSP: fmt.Sprintf("default-src 'self' %[1]s;"+
					"img-src *;"+
					"script-src 'self' %[1]s 'unsafe-eval';"+
					"frame-src 'self' %[2]s;"+
					"style-src 'self' %[1]s 'unsafe-inline';"+
					"connect-src 'self' %[1]s;"+
					"font-src 'self' %[1]s;", r.p.CommonURL(), r.p.PluginURL()),
				View: view,
			}, res)
		}
	} else if mode == "api" {
		// Load the API Views
	}
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
	Tiles       []Tile              `json:"tiles"`
}

type Tile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	View        string `json:"view"`
	Size        string `json:"size"`
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Homepage string `json:"homepage"`
}

type JSONView struct {
	Content []byte
}

func (j *JSONView) Write(w io.Writer) {
	w.Write(j.Content)
}

func (j *JSONView) Code() int           { return 200 }
func (j *JSONView) ContentLength() int  { return len(j.Content) }
func (j *JSONView) ContentType() string { return "application/json" }
func (j *JSONView) Headers() Headers    { return nil }

func LoadPlugins(pluginPath string) View {
	files, err := ioutil.ReadDir(pluginPath)
	if err != nil {
		return &HTTPError{500, "Couldn't read plugin directory."}
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
	bytes, err := json.Marshal(output)
	if err != nil {
		return &HTTPError{500, "Couldn't marshal JSON."}
	}
	return &JSONView{bytes}
}

func WriteView(view View, res http.ResponseWriter) {
	// Set Headers
	res.Header().Add("Content-Type", view.ContentType())
	contentLength := strconv.Itoa(view.ContentLength())
	res.Header().Add("Content-Length", contentLength)

	for key, value := range view.Headers() {
		res.Header().Add(key, value)
	}

	res.WriteHeader(view.Code())

	view.Write(res)
}

type Headers map[string]string

type View interface {
	Write(io.Writer)
	ContentType() string
	ContentLength() int
	Code() int
	Headers() Headers
}

// Http Error
type HTTPError struct {
	ErrorCode int
	Message   string
}

func (h *HTTPError) Write(res io.Writer) {
	res.Write([]byte(h.Message))
}

func (h *HTTPError) ContentType() string {
	return "text/plain"
}

func (h *HTTPError) ContentLength() int {
	return len([]byte(h.Message))
}

func (h *HTTPError) Code() int {
	return h.ErrorCode
}

func (h *HTTPError) Headers() Headers { return nil }

type CSPWrapper struct {
	CSP string
	View
}

func (c *CSPWrapper) Headers() Headers {
	temp := c.View.Headers()
	if temp == nil {
		temp = make(map[string]string)
	}
	temp["Content-Security-Policy"] = c.CSP
	return temp
}

// File View Will Render a File to the Http Response
type FileView struct {
	File *os.File
}

func (f *FileView) Write(w io.Writer) {
	io.Copy(w, f.File)
	f.File.Close()
}

func (b *FileView) ContentLength() int {
	info, _ := b.File.Stat()
	return int(info.Size())
}

func (b *FileView) ContentType() string {
	return mime.TypeByExtension(filepath.Ext(b.File.Name()))
}

func (b *FileView) Code() int {
	return 200
}

func (h *FileView) Headers() Headers { return nil }

// Blatantly stolen from robfig's Revel Framework
func FindFile(prefix string, request string) View {
	var path string

	// Check if Prefix is Absolute, if not prepend the cwd
	if !filepath.IsAbs(prefix) {
		path = os.Getenv("MLGBASE")
	}

	// Get full filename attempted to access
	// Determine whether or not file is in directory
	// TODO: Ensure that Links are not followed
	basePathPrefix := filepath.Join(path, filepath.FromSlash(prefix))
	fname := filepath.Join(basePathPrefix, filepath.FromSlash(request))

	if !strings.HasPrefix(fname, basePathPrefix) {
		return &HTTPError{403, "File not allowed."}
	}

	// Get information on the file
	finfo, err := os.Stat(fname)
	if err != nil {
		// If the file isn't found, return a 404.
		if os.IsNotExist(err) || err.(*os.PathError).Err == syscall.ENOTDIR {
			return &HTTPError{404, "File not found."}
		}
		fmt.Println("Error checking file:", err)
		return &HTTPError{500, "Couldn't get file."}
	}

	// Check if it is a directory listing
	if finfo.Mode().IsDir() {
		return &HTTPError{403, "Directory listing not allowed."}
	}

	// Ensure that we aren't symlinked somewhere terrible
	fqn, err := filepath.EvalSymlinks(fname)
	if err != nil {
		fmt.Println("Error evaling symlinks:", err)
		return &HTTPError{500, "Couldn't eval symlinks."}
	}

	// Open the file for reading
	file, err := os.Open(fqn)
	if err != nil {
		// Check again for existence
		if os.IsNotExist(err) {
			return &HTTPError{404, "File not found."}
		}
		return &HTTPError{500, "Error opening file."}
	}
	return &FileView{file}
}

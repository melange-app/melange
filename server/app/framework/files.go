package framework

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

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
func ServeFile(prefix string, request string) View {
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
		return Error404
	}

	// Get information on the file
	finfo, err := os.Stat(fname)
	if err != nil {
		// If the file isn't found, return a 404.
		if os.IsNotExist(err) || err.(*os.PathError).Err == syscall.ENOTDIR {
			return Error404
		}
		fmt.Println("Error checking file:", err)
		return Error500
	}

	// Check if it is a directory listing
	if finfo.Mode().IsDir() {
		return Error404
	}

	// Ensure that we aren't symlinked somewhere terrible
	fqn, err := filepath.EvalSymlinks(fname)
	if err != nil {
		fmt.Println("Error evaling symlinks:", err)
		return Error500
	}

	// Open the file for reading
	file, err := os.Open(fqn)
	if err != nil {
		// Check again for existence
		if os.IsNotExist(err) {
			return Error404
		}
		return Error500
	}
	return &FileView{file}
}

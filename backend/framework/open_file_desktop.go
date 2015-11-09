// +build !android

package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func GetFile(prefix string, request string) (*FileView, error) {
	var path string

	// Check if Prefix is Absolute, if not prepend the cwd
	if !filepath.IsAbs(prefix) && runtime.GOARCH != "arm" {
		path = os.Getenv("MLGBASE")
	}

	// Get full filename attempted to access
	// Determine whether or not file is in directory
	// TODO: Ensure that Links are not followed
	basePathPrefix := filepath.Join(path, filepath.FromSlash(prefix))
	fname := filepath.Join(basePathPrefix, filepath.FromSlash(request))

	if !strings.HasPrefix(fname, basePathPrefix) {
		fmt.Println("File has incorrect prefix", basePathPrefix)
		return nil, errNoFile
	}

	// Get information on the file
	finfo, err := os.Stat(fname)
	if err != nil {
		// If the file isn't found, return a 404.
		if os.IsNotExist(err) || err.(*os.PathError).Err == syscall.ENOTDIR {
			return nil, errNoFile
		}
		fmt.Println("Error checking file:", err)
		return nil, err
	}

	// Check if it is a directory listing
	if finfo.Mode().IsDir() {
		fmt.Println("Turns out it's a directory!")
		return nil, errNoFile
	}

	// Ensure that we aren't symlinked somewhere terrible
	fqn, err := filepath.EvalSymlinks(fname)
	if err != nil {
		fmt.Println("Error evaling symlinks:", err)
		return nil, err
	}

	// Open the file for reading
	file, err := os.Open(fqn)
	if err != nil {
		// Check again for existence
		if os.IsNotExist(err) {
			return nil, errNoFile
		}
		return nil, err
	}

	info, _ := file.Stat()

	return &FileView{
		File: file,
		Name: file.Name(),
		Size: int(info.Size()),
	}, nil
}

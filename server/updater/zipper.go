package updater

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func cloneZipItem(f *zip.File, dest string) error {
	if strings.HasPrefix(f.Name, "__MACOSX") {
		return nil
	}
	// Create full directory path
	path := filepath.Join(dest, f.Name)
	// fmt.Println("Creating", path)
	err := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	// Clone if item is a file
	rc, err := f.Open()
	if err != nil {
		return err
	}

	if !f.FileInfo().IsDir() {
		// Use os.Create() since Zip don't store file permissions.
		fileCopy, err := os.Create(path)
		if err != nil {
			return err
		}

		// for _, v := range executables {
		// 	if f.Name == v {
		// 		fileCopy.Chmod()
		// 	}
		// }

		_, err = io.Copy(fileCopy, rc)
		fileCopy.Close()
		if err != nil {
			return err
		}
	}
	rc.Close()
	return nil
}

func extractZip(zipPath io.ReaderAt, size int64, dest string) error {
	r, err := zip.NewReader(zipPath, size)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		err = cloneZipItem(f, dest)
		if err != nil {
			return err
		}
	}
	return nil
}

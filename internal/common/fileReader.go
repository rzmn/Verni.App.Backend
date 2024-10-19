package common

import (
	"log"
	"os"
	"path/filepath"
)

var relativePathRoot = func() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	executablePath := filepath.Dir(ex)
	log.Printf("default relative paths root: %s", executablePath)
	return executablePath
}()

func RegisterRelativePathRoot(root string) {
	log.Printf("override relative paths root: %s", root)
	relativePathRoot = root
}

func AbsolutePath(relativePath string) string {
	return filepath.Join(relativePathRoot, relativePath)
}

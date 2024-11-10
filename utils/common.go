package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CreateDirIfNotExists(dirPath string) error {
	return os.MkdirAll(dirPath, 0777)
}

func InitDir(dirPath string) error {
	// Create objects sub-directory
	err := os.Mkdir(filepath.Join(dirPath, "objects"), 0777)
	if err != nil {
		return err
	}

	// Create refs + refs/heads & refs/tags sub-directory
	err = os.Mkdir(filepath.Join(dirPath, "refs"), 0777)
	if err != nil {
		return err
	}
	err = os.Mkdir(filepath.Join(dirPath, "refs", "heads"), 0777)
	if err != nil {
		return err
	}
	err = os.Mkdir(filepath.Join(dirPath, "refs", "tags"), 0777)
	if err != nil {
		return err
	}

	// Create HEAD file -> ref: refs/heads/master
	err = os.WriteFile(filepath.Join(dirPath, "HEAD"), []byte("ref: refs/heads/master\n"), 0777)
	if err != nil {
		return err
	}

	// Create config file:
	// [core]
	// repositoryformatversion = 0
	// filemode = false
	// bare = false
	err = os.WriteFile(filepath.Join(dirPath, "config"), []byte("[core]\nrepositoryformatversion = 0\nfilemode = false\nbare = false"), 0777)
	if err != nil {
		return err
	}

	return nil
}

func GetGoTrackDir(cwd string) (string, error) {
	dirEntries, err := os.ReadDir(cwd)
	if err != nil {
		return "", err
	}

	var gotrackPath string
	for _, entry := range dirEntries {
		if strings.Compare(entry.Name(), ".gotrack") == 0 {
			gotrackPath = filepath.Join(cwd, entry.Name())
			return gotrackPath, nil
		}
	}

	// Recurse one directory above
	if gotrackPath == "" {
		parentDir := filepath.Join(cwd, "..")
		fmt.Println("couldn't find path, moving to", parentDir)
		return GetGoTrackDir(parentDir)
	}

	return "", nil
}

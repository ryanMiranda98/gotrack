package utils

import (
	"crypto/sha1"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	. "github.com/ryanMiranda98/gotrack/internal/errors"
)

func CreateDir(dirPath string) error {
	err := os.MkdirAll(dirPath, fs.ModePerm)
	if err != nil {
		log.Println(err)
		return ErrDirCreationFailed
	}

	return nil
}

func InitDir(dirPath string) error {
	dirs := []string{
		"objects",
		"refs",
		filepath.Join("refs", "heads"),
		filepath.Join("refs", "tags"),
	}

	for _, dir := range dirs {
		err := CreateDir(filepath.Join(dirPath, dir))
		if err != nil {
			return err
		}
	}

	// Create HEAD file -> ref: refs/heads/master
	path := filepath.Join(dirPath, "HEAD")
	err := os.WriteFile(path, []byte("ref: refs/heads/master\n"), fs.ModePerm)
	if err != nil {
		return ErrFileNotWritten
	}

	// Create config file:
	// [core]
	// repositoryformatversion = 0
	// filemode = false
	// bare = false
	path = filepath.Join(dirPath, "config")
	err = os.WriteFile(path, []byte("[core]\nrepositoryformatversion = 0\nfilemode = false\nbare = false"), fs.ModePerm)
	if err != nil {
		return ErrFileNotWritten
	}

	return nil
}

func GetGoTrackDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return RecurseGoTrackWd(cwd)
}

func RecurseGoTrackWd(cwd string) (string, error) {
	dirEntries, err := os.ReadDir(cwd)
	if err != nil {
		return "", ErrReadDir
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
		return RecurseGoTrackWd(parentDir)
	}

	return gotrackPath, nil
}

func GetParentDir(cwd string) string {
	return filepath.Dir(cwd)
}

func GenerateSHA1Hash(data string) (string, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(data))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

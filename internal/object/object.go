package object

import (
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ryanMiranda98/gotrack/utils"
)

type Object struct {
	Type    string
	Size    int
	Content []byte
}

func CreateNewObject(objectType string, size int, content []byte) *Object {
	var object Object
	switch strings.ToLower(objectType) {
	case "blob":
		object.Type = "blob"
		object.Size = size
		object.Content = content
	}

	return &object
}

func (o *Object) Serialize() string {
	return fmt.Sprintf("%s %d\x00%s", o.Type, o.Size, string(o.Content))
}

func (o *Object) Deserialize(data string) error {
	objType, split, found := strings.Cut(data, " ")
	if !found {
		return errors.New("invalid data provided.")
	}

	o.Type = objType
	size, content, found := strings.Cut(split, "\x00")
	if !found {
		return errors.New("invalid data provided.")
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		return err
	}
	o.Size = sizeInt
	o.Content = []byte(content)

	return nil
}

func (o *Object) GenerateSHA1Hash() (string, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(o.Serialize()))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func CreateAndStoreBlob(path string) (string, int, error) {
	filePath := filepath.Clean(path)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", -1, err
	}
	gotrackObject := CreateNewObject("blob", len(content), content)
	hash, err := gotrackObject.GenerateSHA1Hash()
	if err != nil {
		return "", -1, err
	}

	dir := hash[0:2]
	filename := hash[2:]

	cwd, err := os.Getwd()
	if err != nil {
		return "", -1, err
	}

	vcsDir, err := utils.GetGoTrackDir(cwd)
	if err != nil {
		return "", -1, err
	}

	objectsFilePath := filepath.Join(vcsDir, "objects", dir)
	err = utils.CreateDirIfNotExists(objectsFilePath)
	if err != nil {
		return "", -1, err
	}
	contentPath := filepath.Join(objectsFilePath, filename)
	file, err := os.Create(contentPath)
	if err != nil {
		return "", -1, err
	}
	defer file.Close()

	writer := zlib.NewWriter(file)
	defer writer.Close()
	_, err = writer.Write([]byte(gotrackObject.Serialize()))
	if err != nil {
		return "", -1, err
	}
	writer.Flush()

	return hash, len(content), nil
}

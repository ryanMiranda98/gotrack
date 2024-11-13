package object

import (
	"compress/zlib"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/ryanMiranda98/gotrack/internal/errors"
	"github.com/ryanMiranda98/gotrack/utils"
)

// Blob
// blob <size>\x00<content>

type Blob struct {
	Type    string
	Size    int
	Content []byte
}

func CreateNewObject(objectType string, content []byte) Object {
	var object Object
	switch strings.ToLower(objectType) {
	case "blob":
		var blob Blob
		blob.Type = "blob"
		blob.Size = len(content)
		blob.Content = content
		object = &blob
	}

	return object
}

func (b *Blob) Serialize() string {
	return fmt.Sprintf("%s %d\x00%s", b.Type, b.Size, string(b.Content))
}

func (b *Blob) Deserialize(data string) error {
	objType, split, found := strings.Cut(data, " ")
	if !found {
		return ErrInvalidData
	}

	b.Type = objType
	size, content, found := strings.Cut(split, "\x00")
	if !found {
		return ErrInvalidData
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		return err
	}
	b.Size = sizeInt
	b.Content = []byte(content)

	return nil
}

func CreateAndStoreBlob(path string, writeToFileFlag bool) (string, int, error) {
	filePath := filepath.Clean(path)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", -1, err
	}

	gotrackObject := CreateNewObject("blob", content)
	serializedContent := gotrackObject.Serialize()
	hash, err := utils.GenerateSHA1Hash(serializedContent)
	if err != nil {
		return "", -1, err
	}

	dir := hash[0:2]
	filename := hash[2:]

	vcsDir, err := utils.GetGoTrackDir()
	if err != nil {
		return "", -1, err
	}

	objectsFilePath := filepath.Join(vcsDir, "objects", dir)
	err = utils.CreateDir(objectsFilePath)
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
	_, err = writer.Write([]byte(serializedContent))
	if err != nil {
		return "", -1, err
	}
	writer.Flush()

	return hash, len(content), nil
}

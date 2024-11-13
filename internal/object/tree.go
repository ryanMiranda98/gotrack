package object

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	. "github.com/ryanMiranda98/gotrack/internal/errors"
	"github.com/ryanMiranda98/gotrack/utils"
)

// Tree
// tree <size>
// <mode> <name>\x00<hash>

const (
	MODE_BLOB    = "100644"
	MODE_TREE    = "040000"
	MODE_SYMLINK = "120000"
)

type Tree struct {
	Type        string
	Size        int
	TreeEntries []TreeEntry
}

type TreeEntry struct {
	Mode string
	Name string
	Hash string
	Size int
}

func CreateNewTree(treeEntries *[]TreeEntry) *Tree {
	var object Tree
	object.Type = "tree"
	object.TreeEntries = *treeEntries

	return &object
}

func (t *Tree) Serialize() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s %d\n", t.Type, t.Size))
	for _, entry := range t.TreeEntries {
		buf.WriteString(fmt.Sprintf("%s %s\x00%s\n", entry.Mode, entry.Name, entry.Hash))
	}
	return buf.String()
}

func (t *Tree) Deserialize(data string) error {
	entries := strings.Split(data, "\n")
	if len(entries) < 2 {
		return ErrInvalidArguments
	}

	var treeEntries []TreeEntry
	entries = entries[1 : len(entries)-1]
	for _, entry := range entries {
		mode, split, found := strings.Cut(entry, " ")
		if !found {
			return ErrInvalidData
		}

		name, hash, found := strings.Cut(split, "\x00")
		if !found {
			return ErrInvalidData
		}

		treeEntries = append(treeEntries, TreeEntry{
			Hash: hash,
			Mode: mode,
			Name: name,
			Size: 0,
		})
	}

	t.TreeEntries = treeEntries
	return nil
}

func CreateAndStoreTree(treeEntries *[]TreeEntry) (string, int, error) {
	tree := CreateNewTree(treeEntries)
	for _, entries := range *treeEntries {
		tree.Size += entries.Size
	}
	hash, err := utils.GenerateSHA1Hash(tree.Serialize())
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
	_, err = writer.Write([]byte(tree.Serialize()))
	if err != nil {
		return "", -1, err
	}
	writer.Flush()

	return hash, tree.Size, nil
}

func GenerateTreeEntries(cwd string, pathsToIgnore *[]string) (string, int, error) {
	dirEntries, err := os.ReadDir(cwd)
	if err != nil {
		return "", -1, err
	}

	var treeEntries []TreeEntry
	for _, entry := range dirEntries {
		// Ignore files from pathsToIgnore
		currentFilePath := filepath.Join(cwd, entry.Name())
		if contains := slices.Contains(*pathsToIgnore, currentFilePath); !contains {
			// For each dir, recursively create a tree object and record its hash
			var hash, mode string
			var size int
			if entry.IsDir() {
				hash, size, err = GenerateTreeEntries(currentFilePath, pathsToIgnore)
				if err != nil {
					return "", -1, err
				}
				mode = MODE_TREE
			} else {
				// For each file/blob, compute hash, store object and return hash and size to tree
				hash, size, err = CreateAndStoreBlob(currentFilePath, true)
				if err != nil {
					return "", -1, err
				}
				mode = MODE_BLOB
			}

			fi, err := os.Stat(currentFilePath)
			if err != nil {
				return "", -1, err
			}
			if fi.Mode() == os.ModeSymlink {
				mode = MODE_SYMLINK
			}
			treeEntries = append(treeEntries, TreeEntry{
				Hash: hash,
				Size: size,
				Mode: mode,
				Name: entry.Name(),
			})
		}
	}

	// Sort TreeEntries by name
	sort.Slice(treeEntries, func(i, j int) bool {
		return treeEntries[i].Name < treeEntries[j].Name
	})

	hash, _, err := CreateAndStoreTree(&treeEntries)
	return hash, -1, err
}

func ListTree(hash, parentDir string, recursive bool) error {
	dir := hash[0:2]
	filename := hash[2:]

	vcsDir, err := utils.GetGoTrackDir()
	if err != nil {
		return err
	}
	treePath := filepath.Join(vcsDir, "objects", dir, filename)
	file, err := os.OpenFile(treePath, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotExists
		}
	}
	defer file.Close()

	// Decompress zlib
	reader, err := zlib.NewReader(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return err
	}

	var tree Tree
	tree.Deserialize(buf.String())

	for _, entry := range tree.TreeEntries {
		var entryType string
		switch entry.Mode {
		case MODE_TREE:
			if recursive {
				var name string
				if parentDir == "" {
					name = entry.Name
				} else {
					name = filepath.Join(parentDir, entry.Name)
				}
				ListTree(entry.Hash, name, true)
				continue
			} else {
				entryType = "tree"
			}
		case MODE_BLOB:
			entryType = "blob"
		}

		var name string
		if parentDir != "" {
			name = filepath.Join(parentDir, entry.Name)
		} else {
			name = entry.Name
		}
		fmt.Printf("%s\t%s\t%s\t%s\n", entry.Mode, entryType, entry.Hash, name)
	}

	return nil
}

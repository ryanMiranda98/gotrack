package command

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/ryanMiranda98/gotrack/internal/errors"
	"github.com/ryanMiranda98/gotrack/internal/object"
	"github.com/ryanMiranda98/gotrack/utils"
	"github.com/spf13/cobra"
)

var (
	writeHashObjectFlag bool
	recursiveLsTreeFlag   bool
)

var rootCmd = &cobra.Command{
	Use:  "gotrack",
	Long: "A simple VCS CLI in Golang. Complete documentation is available at xyz.\nUsage: gotrack <command> [options] [file]",
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(hashObjectCmd)
	rootCmd.AddCommand(catFileCmd)
	rootCmd.AddCommand(writeTreeCmd)
	rootCmd.AddCommand(lsTreeCmd)

	hashObjectCmd.Flags().BoolVarP(&writeHashObjectFlag, "write", "w", false, "compute hash from content and write to file.")
	lsTreeCmd.Flags().BoolVarP(&recursiveLsTreeFlag, "recursive", "r", false, "list tree content recursively.")
}

func Execute() error {
	return rootCmd.Execute()
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initializes a new, empty gotrack repository locally.",
	RunE: func(c *cobra.Command, args []string) error {
		var wd string
		argLength := len(args)

		switch argLength {
		case 0:
			wd = "."
		case 1:
			wd = args[0]
		default:
			return ErrInvalidArguments
		}

		dir := filepath.Join(wd, ".gotrack")

		if dirInfo, err := os.Stat(dir); err == nil {
			if !dirInfo.IsDir() {
				return ErrDirNotExists
			}

			if dirEntries, err := os.ReadDir(dir); err != nil {
				return err
			} else if len(dirEntries) != 0 {
				return ErrDirAlreadyInitialized
			} else {
				return utils.InitDir(dir)
			}
		} else {
			if err := utils.CreateDir(dir); err != nil {
				return err
			}
			return utils.InitDir(dir)
		}
	},
}

var hashObjectCmd = &cobra.Command{
	Use:   "hash-object",
	Short: "Creates a gotrack blob.",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) < 1 {
			return ErrFileArgNotProvided
		}
		hash, _, err := object.CreateAndStoreBlob(args[0], writeHashObjectFlag)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", hash)

		return nil
	},
}

var catFileCmd = &cobra.Command{
	Use:   "cat-file <object_type> <object_hash>",
	Short: "Read the contents of a gotrack blob.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return ErrInvalidArguments
		}

		hash := args[1]

		gotrackDir, err := utils.GetGoTrackDir()
		if err != nil {
			return err
		}

		filePath := filepath.Join(gotrackDir, "objects", hash[0:2], hash[2:])
		_, err = os.Stat(filePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return ErrFileNotExists
			}
		}
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Decompress file
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

		var gotrackObject object.Blob
		gotrackObject.Deserialize(buf.String())
		fmt.Printf("%s\n", string(gotrackObject.Content))
		return nil
	},
}

var lsTreeCmd = &cobra.Command{
	Use:   "ls-tree [-r] <tree_SHA>",
	Short: "Prints the contents of a tree object.",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) != 1 {
			return ErrInvalidArguments
		}
		treeHash := args[0]
		object.ListTree(treeHash, "", recursiveLsTreeFlag)
		return nil
	},
}

// TODO: Implement a staging area. For now, assume all files are staged.
var writeTreeCmd = &cobra.Command{
	Use:   "write-tree",
	Short: "Creates a tree object from the staging area.",
	RunE: func(c *cobra.Command, args []string) error {
		vcsDir, err := utils.GetGoTrackDir()
		if err != nil {
			return err
		}

		// TODO: read more paths from .gotrackignore if exists
		// Add regexp for paths like bin/*
		pathsToIgnore := []string{
			vcsDir,
		}

		cwd := utils.GetParentDir(vcsDir)
		hash, _, err := object.GenerateTreeEntries(cwd, &pathsToIgnore)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", hash)
		return nil
	},
}

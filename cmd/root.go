package command

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ryanMiranda98/gotrack/internal/object"
	"github.com/ryanMiranda98/gotrack/internal/tree"
	"github.com/ryanMiranda98/gotrack/utils"
	"github.com/spf13/cobra"
)

var ErrInvalidArguments = errors.New("invalid argument(s) provided.")
var ErrDirNotExists = errors.New("directory does not exist.")
var ErrFileNotExists = errors.New("file does not exist.")
var ErrDirInitialized = errors.New("directory has already been initialized.")
var ErrFileNotProvided = errors.New("please provide a file in the arguments")

var (
	hashObjectWriteToFileFlag bool
	listTreeRecursivelyFlag bool
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

	hashObjectCmd.Flags().BoolVarP(&hashObjectWriteToFileFlag, "write", "w", false, "compute hash from content and write to file.")
	lsTreeCmd.Flags().BoolVarP(&listTreeRecursivelyFlag, "recursive", "r", false, "list tree content recursively.")
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
				return ErrDirInitialized
			} else {
				return utils.InitDir(dir)
			}
		} else {
			if err := utils.CreateDirIfNotExists(dir); err != nil {
				return err
			}
			return utils.InitDir(dir)
		}
	},
}

var hashObjectCmd = &cobra.Command{
	Use:   "hash-object",
	Short: "Creates a gotrack (git) object.",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) < 1 {
			return ErrFileNotProvided
		}
		if hashObjectWriteToFileFlag {
			hash, _, err := object.CreateAndStoreBlob(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", hash)
		} else {
			filePath := filepath.Clean(args[0])
			content, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			gotrackObject := object.CreateNewObject("blob", len(content), content)
			hash, err := gotrackObject.GenerateSHA1Hash()
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", hash)
		}

		return nil
	},
}

var catFileCmd = &cobra.Command{
	Use:   "cat-file <object_type> <object_hash>",
	Short: "Read the contents of a gotrack object.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return ErrInvalidArguments
		}

		// objectType := args[0]
		hash := args[1]

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		gotrackDir, err := utils.GetGoTrackDir(cwd)
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

		var gotrackObject object.Object
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
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		tree.ListTree(treeHash, cwd, "", listTreeRecursivelyFlag)
		return nil
	},
}

// TODO: Implement a staging area. For now, assume all files are staged.
var writeTreeCmd = &cobra.Command{
	Use:   "write-tree",
	Short: "Creates a tree object from the staging area.",
	RunE: func(c *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		gotrackDir, err := utils.GetGoTrackDir(cwd)
		if err != nil {
			return err
		}

		// TODO: read more paths from .gotrackignore if exists
		// Add regexp for paths like bin/*
		pathsToIgnore := []string{
			gotrackDir,
		}

		hash, _, err := tree.GenerateTreeEntries(cwd, &pathsToIgnore)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", hash)
		return nil
	},
}

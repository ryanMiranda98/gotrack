package errors

import "errors"

var (
	ErrDirCreationFailed     = errors.New("failed to create directory.")
	ErrReadDir               = errors.New("an error occurred while reading directory.")
	ErrDirNotExists          = errors.New("directory does not exist.")
	ErrDirAlreadyInitialized = errors.New("directory has already been initialized.")

	ErrFileCreationFailed = errors.New("failed to create file.")
	ErrFileNotWritten     = errors.New("an error occurred while writing to file.")
	ErrFileNotExists      = errors.New("file does not exist.")
	ErrFileArgNotProvided = errors.New("please provide a file name in the arguments.")
	ErrInvalidArguments   = errors.New("invalid argument(s) provided.")
	ErrInvalidData        = errors.New("invalid data provided.")
)

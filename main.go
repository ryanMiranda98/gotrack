package main

import (
	"os"

	command "github.com/ryanMiranda98/gotrack/cmd"
)

func main() {
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

package main

import (
	"os"

	"github.com/MagnumGOYB/memforge/internal/cli"
)

func main() {
	os.Exit(cli.Execute(os.Args[1:], cli.Streams{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}))
}

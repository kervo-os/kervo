package cli

import "fmt"

// Version is stamped by GoReleaser via -ldflags at release time.
var Version = "dev"

func runVersion(args []string) error {
	fmt.Println("kervo", Version)
	return nil
}

package cli

import (
	"fmt"
	"runtime/debug"
)

// Version is stamped by GoReleaser via -ldflags at release time.
var Version = "dev"

func runVersion(args []string) error {
	fmt.Println("kervo", resolveVersion())
	return nil
}

// resolveVersion falls back to the module version Go embeds at install
// time, so a plain `go install ...@v0.1.0` binary identifies itself
// without GoReleaser.
func resolveVersion() string {
	if Version != "dev" {
		return Version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return Version
}

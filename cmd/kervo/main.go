// Command kervo is the single entry point (ARCH-0001 §1.3: daemonless CLI).
package main

import (
	"fmt"
	"os"

	"github.com/kervo-os/kervo/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "kervo:", err)
		os.Exit(1)
	}
}

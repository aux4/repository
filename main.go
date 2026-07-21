package main

import (
	"fmt"
	"os"
)

// exitInjection is the exit code used for empty results, rejected expressions,
// and SQL parse errors, matching the pure-aux4 implementation it replaces.
const exitInjection = 4

func arg(args []string, i int) string {
	if i < len(args) {
		return args[i]
	}
	return ""
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: missing command (expected init, write, read, find, count, patch, touch, delete, truncate or clean)")
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "init":
		err = cmdInit(args)
	case "write":
		err = cmdWrite(args)
	case "read":
		err = cmdRead(args)
	case "find":
		err = cmdFind(args)
	case "count":
		err = cmdCount(args)
	case "patch":
		err = cmdPatch(args)
	case "touch":
		err = cmdTouch(args)
	case "delete":
		err = cmdDelete(args)
	case "truncate":
		err = cmdTruncate(args)
	case "clean":
		err = cmdClean(args)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", cmd)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

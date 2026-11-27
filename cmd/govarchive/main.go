package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, out, diagnostics io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: govarchive <capture|verify|diff|verify-upgrade> [flags]")
	}
	switch args[0] {
	case "capture":
		return runCapture(args[1:], out, diagnostics)
	case "verify":
		return runVerify(args[1:], out, diagnostics)
	case "diff":
		return runDiff(args[1:], out, diagnostics)
	case "verify-upgrade":
		return verifyUpgrade(args[1:], out, diagnostics)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

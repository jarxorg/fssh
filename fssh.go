package fssh

import (
	"flag"
	"fmt"
)

// Main runs shell.
func Main(osArgs []string) error {
	flagSet := flag.NewFlagSet(ShellName, flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Printf("Usage:\n  %s ([dir])\n", ShellName)
		fmt.Println("Examples:")
		fmt.Printf("  %s\n", ShellName)
		fmt.Printf("  %s DIR\n", ShellName)
		fmt.Printf("  %s (s3|gs)://BUCKET/\n", ShellName)
	}
	if err := flagSet.Parse(osArgs[1:]); err != nil {
		return err
	}

	args := flagSet.Args()
	dirUrl := "."
	if len(args) > 0 {
		dirUrl = args[0]
	}
	sh, err := NewShell(dirUrl)
	if err != nil {
		return err
	}
	defer sh.Close()

	return sh.Run()
}

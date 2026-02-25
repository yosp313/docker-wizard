package main

import (
	"flag"
	"fmt"
	"os"

	"docker-wizard/internal/app"
)

var version = "dev"

func main() {
	showVersion, mode, err := parseArgs(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		printUsage()
		os.Exit(2)
	}

	if showVersion {
		fmt.Printf("docker-wizard:%s\n", version)
		return
	}

	if err := app.RunWithOptions(app.Options{Mode: mode}); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func parseArgs(args []string) (bool, app.Mode, error) {
	fs := flag.NewFlagSet("docker-wizard", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	modeFlag := fs.String("mode", string(app.ModeStyled), "run mode: styled, plain, cli")
	versionFlag := fs.Bool("version", false, "print version")
	versionShortFlag := fs.Bool("v", false, "print version")

	if err := fs.Parse(args); err != nil {
		return false, "", err
	}
	if fs.NArg() > 0 {
		return false, "", fmt.Errorf("unexpected arguments: %v", fs.Args())
	}

	mode, err := app.ParseMode(*modeFlag)
	if err != nil {
		return false, "", err
	}

	return *versionFlag || *versionShortFlag, mode, nil
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: docker-wizard [--mode styled|plain|cli] [--version]")
}

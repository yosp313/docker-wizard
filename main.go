package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"docker-wizard/internal/app"
)

var version = "dev"

func main() {
	showVersion, options, err := parseArgs(os.Args[1:])
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

	if err := app.RunWithOptions(options); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func parseArgs(args []string) (bool, app.Options, error) {
	fs := flag.NewFlagSet("docker-wizard", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	modeFlag := fs.String("mode", string(app.ModeStyled), "run mode: styled, plain, cli, batch")
	servicesFlag := fs.String("services", "", "comma-separated service IDs (batch mode)")
	languageFlag := fs.String("language", "", "language override: go, node, python, ruby, php, java, dotnet (batch mode)")
	writeFlag := fs.Bool("write", false, "write generated files (batch mode)")
	dryRunFlag := fs.Bool("dry-run", false, "preview only; do not write files (batch mode)")
	versionFlag := fs.Bool("version", false, "print version")
	versionShortFlag := fs.Bool("v", false, "print version")

	if err := fs.Parse(args); err != nil {
		return false, app.Options{}, err
	}
	if fs.NArg() > 0 {
		return false, app.Options{}, fmt.Errorf("unexpected arguments: %v", fs.Args())
	}

	mode, err := app.ParseMode(*modeFlag)
	if err != nil {
		return false, app.Options{}, err
	}

	usesAutomationFlags := strings.TrimSpace(*servicesFlag) != "" || strings.TrimSpace(*languageFlag) != "" || *writeFlag || *dryRunFlag
	if mode != app.ModeBatch && usesAutomationFlags {
		return false, app.Options{}, fmt.Errorf("--services, --language, --write, and --dry-run require --mode batch")
	}
	if mode == app.ModeBatch && *writeFlag && *dryRunFlag {
		return false, app.Options{}, fmt.Errorf("--write and --dry-run cannot be used together")
	}

	return *versionFlag || *versionShortFlag, app.Options{
		Mode: mode,
		Automation: app.AutomationOptions{
			Services: parseServicesFlag(*servicesFlag),
			Language: strings.TrimSpace(*languageFlag),
			Write:    *writeFlag,
			DryRun:   *dryRunFlag,
		},
	}, nil
}

func parseServicesFlag(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	tokens := strings.Split(value, ",")
	services := make([]string, 0, len(tokens))
	seen := map[string]bool{}
	for _, token := range tokens {
		normalized := strings.ToLower(strings.TrimSpace(token))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		services = append(services, normalized)
	}

	return services
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: docker-wizard [--mode styled|plain|cli|batch] [--version]")
	fmt.Fprintln(os.Stderr, "batch mode flags: --services mysql,redis --language go [--write|--dry-run]")
}

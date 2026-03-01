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
	// check for subcommands before flag parsing
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "add":
			runAdd(os.Args[2:])
			return
		case "list":
			runList()
			return
		}
	}

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
	if *versionFlag || *versionShortFlag {
		return true, app.Options{}, nil
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

	return false, app.Options{
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

func runAdd(args []string) {
	fs := flag.NewFlagSet("docker-wizard add", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	writeFlag := fs.Bool("write", false, "apply changes (default is dry-run)")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		printAddUsage()
		os.Exit(2)
	}

	serviceIDs := fs.Args()
	if len(serviceIDs) == 0 {
		fmt.Fprintln(os.Stderr, "error: at least one service ID is required")
		printAddUsage()
		os.Exit(2)
	}

	// normalize service IDs
	normalized := make([]string, 0, len(serviceIDs))
	seen := map[string]bool{}
	for _, id := range serviceIDs {
		id = strings.ToLower(strings.TrimSpace(id))
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		normalized = append(normalized, id)
	}

	if err := app.RunAdd(app.AddOptions{
		Services: normalized,
		Write:    *writeFlag,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runList() {
	if err := app.RunList(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: docker-wizard [command] [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  add <service...>  add services to existing compose file")
	fmt.Fprintln(os.Stderr, "  list              show available services")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "wizard flags:")
	fmt.Fprintln(os.Stderr, "  --mode styled|plain|cli|batch")
	fmt.Fprintln(os.Stderr, "  --version")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "batch mode flags:")
	fmt.Fprintln(os.Stderr, "  --services mysql,redis --language go [--write|--dry-run]")
}

func printAddUsage() {
	fmt.Fprintln(os.Stderr, "usage: docker-wizard add [--write] <service...>")
	fmt.Fprintln(os.Stderr, "  preview by default; pass --write to apply changes")
}

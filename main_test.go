package main

import (
	"testing"

	"docker-wizard/internal/app"
)

func TestParseArgsBatchMode(t *testing.T) {
	showVersion, options, err := parseArgs([]string{"--mode", "batch", "--services", "mysql, redis,mysql", "--language", "go", "--write"})
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}
	if showVersion {
		t.Fatal("expected showVersion false")
	}
	if options.Mode != app.ModeBatch {
		t.Fatalf("expected mode %q, got %q", app.ModeBatch, options.Mode)
	}
	if got, want := options.Automation.Language, "go"; got != want {
		t.Fatalf("expected language %q, got %q", want, got)
	}
	if !options.Automation.Write {
		t.Fatal("expected write true")
	}
	if options.Automation.DryRun {
		t.Fatal("expected dry-run false")
	}

	gotServices := options.Automation.Services
	wantServices := []string{"mysql", "redis"}
	if len(gotServices) != len(wantServices) {
		t.Fatalf("expected services %v, got %v", wantServices, gotServices)
	}
	for i := range gotServices {
		if gotServices[i] != wantServices[i] {
			t.Fatalf("expected services %v, got %v", wantServices, gotServices)
		}
	}
}

func TestParseArgsRejectsAutomationFlagsOutsideBatch(t *testing.T) {
	_, _, err := parseArgs([]string{"--mode", "cli", "--services", "mysql"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseArgsRejectsWriteAndDryRun(t *testing.T) {
	_, _, err := parseArgs([]string{"--mode", "batch", "--write", "--dry-run"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseArgsVersionIgnoresOtherFlags(t *testing.T) {
	showVersion, _, err := parseArgs([]string{"--version", "--services", "mysql"})
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}
	if !showVersion {
		t.Fatal("expected showVersion true")
	}
}

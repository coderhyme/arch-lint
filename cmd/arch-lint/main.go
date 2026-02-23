package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/coderhyme/arch-lint/internal/checker"
	"github.com/coderhyme/arch-lint/internal/config"
	"github.com/coderhyme/arch-lint/internal/groups"
	"github.com/coderhyme/arch-lint/internal/loader"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", ".arch-lint.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	manager, err := groups.NewGroupManager(cfg)
	if err != nil {
		log.Fatalf("Failed to build group manager: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	modulePath, packages, errs := loader.Load(cwd)
	if len(errs) > 0 {
		for _, e := range errs {
			log.Printf("Warning: %v", e)
		}
	}
	if modulePath == "" {
		log.Fatalf("Failed to determine module path")
	}

	ctx := context.Background()
	result, err := checker.Check(ctx, modulePath, packages, manager)
	if err != nil {
		log.Fatalf("Failed to check dependencies: %v", err)
	}

	if len(result.Violations) == 0 {
		fmt.Printf("No violations found (%d packages checked)\n", result.PackagesCount)
		return
	}

	fmt.Printf("Found %d violation(s):\n\n", len(result.Violations))
	for _, v := range result.Violations {
		fmt.Printf("  %s\n    imports %s\n    denied by group %q\n\n", v.Package, v.Import, v.GroupName)
	}

	os.Exit(1)
}

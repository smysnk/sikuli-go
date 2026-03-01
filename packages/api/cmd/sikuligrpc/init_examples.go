package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed examples/node examples/python
var embeddedExamples embed.FS

func maybeRunInitExamples(args []string) (bool, error) {
	if len(args) == 0 {
		return false, nil
	}
	if args[0] != "init-examples" {
		return false, nil
	}
	return true, runInitExamples(args[1:])
}

func runInitExamples(args []string) error {
	fs := flag.NewFlagSet("init-examples", flag.ContinueOnError)
	lang := fs.String("lang", "node", "example language: node|javascript|python")
	targetDir := fs.String("dir", ".", "target directory")
	force := fs.Bool("force", false, "overwrite existing examples directory")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: sikuligo init-examples [--lang node|python] [--dir <targetDir>] [--force]\n")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	normalizedLang, sourceRoot, err := resolveExampleLanguage(*lang)
	if err != nil {
		return err
	}
	outputExamplesDir := filepath.Join(*targetDir, "examples")
	if stat, err := os.Stat(outputExamplesDir); err == nil && stat.IsDir() {
		if !*force {
			return fmt.Errorf("target already exists: %s (use --force to overwrite)", outputExamplesDir)
		}
		if err := os.RemoveAll(outputExamplesDir); err != nil {
			return fmt.Errorf("remove existing target: %w", err)
		}
	}
	if err := copyEmbeddedExamples(sourceRoot, outputExamplesDir); err != nil {
		return err
	}

	fmt.Printf("Initialized %s examples in: %s\n", normalizedLang, outputExamplesDir)
	return nil
}

func resolveExampleLanguage(lang string) (string, string, error) {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "node", "javascript", "js":
		return "node", "examples/node", nil
	case "python", "py":
		return "python", "examples/python", nil
	default:
		return "", "", fmt.Errorf("unsupported language %q (expected node|python)", lang)
	}
}

func copyEmbeddedExamples(sourceRoot, outputDir string) error {
	return fs.WalkDir(embeddedExamples, sourceRoot, func(entryPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel := strings.TrimPrefix(entryPath, sourceRoot)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return nil
		}
		target := filepath.Join(outputDir, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		content, err := embeddedExamples.ReadFile(entryPath)
		if err != nil {
			return fmt.Errorf("read embedded example %s: %w", entryPath, err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create output dir for %s: %w", target, err)
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return fmt.Errorf("write output file %s: %w", target, err)
		}
		return nil
	})
}

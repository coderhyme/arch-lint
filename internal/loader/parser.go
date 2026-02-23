package loader

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

type repoTraverser struct {
	rootPath       string
	errs           []error
	packageImports map[string]map[string]struct{}
}

func Load(rootPath string) (modulePath string, packages map[string]map[string]struct{}, errs []error) {
	rt := &repoTraverser{
		rootPath:       rootPath,
		packageImports: make(map[string]map[string]struct{}),
	}

	rt.traverse()

	goModPath := filepath.Join(rootPath, "go.mod")
	file, err := os.ReadFile(goModPath)
	if err != nil {
		return "", nil, append(rt.errs, fmt.Errorf("failed to read go.mod: %w", err))
	}

	mod, err := modfile.ParseLax(goModPath, file, nil)
	if err != nil {
		return "", nil, append(rt.errs, fmt.Errorf("failed to parse go.mod: %w", err))
	}

	return mod.Module.Mod.Path, rt.packageImports, rt.errs
}

var errDoNotSkip = errors.New("do not skip")

func (*repoTraverser) shouldSkip(path string, d fs.DirEntry) error {
	if d.IsDir() {
		if strings.HasPrefix(d.Name(), ".") || d.Name() == "vendor" {
			return filepath.SkipDir
		}
		return nil
	}

	if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
		return nil
	}
	return errDoNotSkip
}

func (rt *repoTraverser) updateImports(path string, imports []string) {
	dir := filepath.Dir(path)
	relDir, err := filepath.Rel(rt.rootPath, dir)
	if err != nil {
		rt.errs = append(rt.errs, fmt.Errorf("failed to get relative path for %s: %w", dir, err))
		return
	}

	if _, exists := rt.packageImports[relDir]; !exists {
		rt.packageImports[relDir] = make(map[string]struct{})
	}

	for _, imp := range imports {
		rt.packageImports[relDir][imp] = struct{}{}
	}
}

func (rt *repoTraverser) traverse() {
	_ = filepath.WalkDir(rt.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			rt.errs = append(rt.errs, err)
			return nil
		}

		if err := rt.shouldSkip(path, d); err != errDoNotSkip {
			return err
		}

		imports, err := extractImports(path)
		if err != nil {
			rt.errs = append(rt.errs, fmt.Errorf("failed to extract imports from %s: %w", path, err))
			return nil
		}

		rt.updateImports(path, imports)

		return nil
	})
}

func extractImports(filename string) ([]string, error) {
	node, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var imports []string
	for _, imp := range node.Imports {
		imports = append(imports, strings.Trim(imp.Path.Value, `"`))
	}

	return imports, nil
}

package checker

import (
	"context"
	"strings"

	"github.com/coderhyme/arch-lint/internal/groups"
)

type Violation struct {
	Package   string
	Import    string
	GroupName string
}

type Result struct {
	Violations    []Violation
	PackagesCount int
}

func Check(ctx context.Context, modulePath string, packageImports map[string]map[string]struct{}, manager groups.GroupManager) (*Result, error) {
	result := &Result{
		PackagesCount: len(packageImports),
	}

	for pkgPath, imports := range packageImports {
		matchingGroups, err := manager.GetGroups(ctx, pkgPath)
		if err != nil {
			return nil, err
		}

		if len(matchingGroups) == 0 {
			continue
		}

		for imp := range imports {
			relImport, ok := stripModulePrefix(modulePath, imp)
			if !ok {
				continue
			}

			for _, grp := range matchingGroups {
				checker := grp.GetDependencyChecker(pkgPath)
				if !checker.CanDependOn(relImport) {
					result.Violations = append(result.Violations, Violation{
						Package:   pkgPath,
						Import:    relImport,
						GroupName: grp.Name(),
					})
				}
			}
		}
	}

	return result, nil
}

func stripModulePrefix(modulePath, importPath string) (string, bool) {
	if !strings.HasPrefix(importPath, modulePath+"/") {
		return "", false
	}
	return strings.TrimPrefix(importPath, modulePath+"/"), true
}

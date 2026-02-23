package groups

import "context"

type GroupManager interface {
	GetGroups(ctx context.Context, path string) ([]Group, error)
	GetGroup(ctx context.Context, name string) (Group, error)
}

type DependencyChecker interface {
	CanDependOn(importPath string) bool
}

type Group interface {
	Name() string
	MatchPath(path string) bool
	GetDependencyChecker(path string) DependencyChecker
}

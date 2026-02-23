package groups

import (
	"context"

	"github.com/coderhyme/arch-lint/internal/config"
)

func newGroup(ctx context.Context, name string, cfg *config.Group, manager GroupManager) (Group, error) {
	pathMatchers := buildPathMatchers(cfg.Paths)

	var denyRules []ImportRule
	var allowRules []ImportRule

	if cfg.Dependencies != nil {
		if cfg.Dependencies.Deny != nil {
			rules, err := buildDependencyMatchers(ctx, cfg.Dependencies.Deny, manager)
			if err != nil {
				return nil, err
			}
			denyRules = append(denyRules, rules...)
		}
		if cfg.Dependencies.Allow != nil {
			rules, err := buildDependencyMatchers(ctx, cfg.Dependencies.Allow, manager)
			if err != nil {
				return nil, err
			}
			allowRules = append(allowRules, rules...)
		}
	}

	return &groupWithRules{
		name:         name,
		pathMatchers: pathMatchers,
		denyRules:    denyRules,
		allowRules:   allowRules,
	}, nil
}

func buildPathMatchers(paths config.PathConfigs) []Matcher {
	var matchers []Matcher
	for _, path := range paths {
		matchers = append(matchers, NewGlobMatcher(path.Dir))
	}
	return matchers
}

func buildDependencyMatchers(ctx context.Context, rule *config.DependencyRule, manager GroupManager) ([]ImportRule, error) {
	var matchers []ImportRule

	for _, pattern := range rule.Patterns {
		matchers = append(matchers, NewGlobImportRule(pattern))
	}

	for _, rel := range rule.Relative {
		matchers = append(matchers, NewRelativeImportRule(rel))
	}

	for _, groupName := range rule.Groups {
		grp, err := manager.GetGroup(ctx, groupName)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, NewGroupImportRule(grp))
	}

	if rule.SubPackages {
		matchers = append(matchers, NewSubPackageImportRule())
	}

	return matchers, nil
}

type groupWithRules struct {
	name         string
	pathMatchers []Matcher
	denyRules    []ImportRule
	allowRules   []ImportRule
}

func (p *groupWithRules) Name() string {
	return p.name
}

func (p *groupWithRules) MatchPath(path string) bool {
	for _, matcher := range p.pathMatchers {
		if matcher.Match(path) {
			return true
		}
	}
	return false
}

func (p *groupWithRules) GetDependencyChecker(path string) DependencyChecker {
	return &ruleBasedChecker{
		packagePath: path,
		denyRules:   p.denyRules,
		allowRules:  p.allowRules,
	}
}

type ruleBasedChecker struct {
	packagePath string
	denyRules   []ImportRule
	allowRules  []ImportRule
}

func (r *ruleBasedChecker) CanDependOn(importPath string) bool {
	if len(r.denyRules) == 0 && len(r.allowRules) == 0 {
		return true
	}

	for _, rule := range r.denyRules {
		if rule.Allows(r.packagePath, importPath) {
			return false
		}
	}

	for _, rule := range r.allowRules {
		if rule.Allows(r.packagePath, importPath) {
			return true
		}
	}

	return false
}

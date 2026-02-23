package checker

import (
	"context"
	"testing"

	"github.com/coderhyme/arch-lint/internal/config"
	"github.com/coderhyme/arch-lint/internal/groups"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CheckerSuite struct {
	suite.Suite
}

func TestCheckerSuite(t *testing.T) {
	suite.Run(t, new(CheckerSuite))
}

func (s *CheckerSuite) TestNoViolation_AllowedGroupDependency() {
	// given
	cfg := &config.Config{
		Version: 1,
		Groups: map[string]*config.Group{
			"domain": {
				Paths: config.PathConfigs{{Dir: "internal/domain/**"}},
			},
			"service": {
				Paths: config.PathConfigs{{Dir: "internal/service/**"}},
				Dependencies: &config.Dependencies{
					Allow: &config.DependencyRule{
						Groups: []string{"domain"},
					},
				},
			},
		},
	}
	manager, err := groups.NewGroupManager(cfg)
	s.Require().NoError(err)

	packages := map[string]map[string]struct{}{
		"internal/service/user": {
			"github.com/example/app/internal/domain/model": {},
		},
	}

	// when
	result, err := Check(context.Background(), "github.com/example/app", packages, manager)

	// then
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), result.Violations)
	assert.Equal(s.T(), 1, result.PackagesCount)
}

func (s *CheckerSuite) TestViolation_DeniedGroupDependency() {
	// given
	cfg := &config.Config{
		Version: 1,
		Groups: map[string]*config.Group{
			"repository": {
				Paths: config.PathConfigs{{Dir: "internal/repository/**"}},
			},
			"domain": {
				Paths: config.PathConfigs{{Dir: "internal/domain/**"}},
				Dependencies: &config.Dependencies{
					Deny: &config.DependencyRule{
						Groups: []string{"repository"},
					},
				},
			},
		},
	}
	manager, err := groups.NewGroupManager(cfg)
	s.Require().NoError(err)

	packages := map[string]map[string]struct{}{
		"internal/domain/user": {
			"github.com/example/app/internal/repository/db": {},
		},
	}

	// when
	result, err := Check(context.Background(), "github.com/example/app", packages, manager)

	// then
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Violations, 1)
	assert.Equal(s.T(), "internal/domain/user", result.Violations[0].Package)
	assert.Equal(s.T(), "internal/repository/db", result.Violations[0].Import)
	assert.Equal(s.T(), "domain", result.Violations[0].GroupName)
}

func (s *CheckerSuite) TestDenyTakesPrecedenceOverAllow() {
	// given - deny "repository" group, but also allow pattern that would match it
	cfg := &config.Config{
		Version: 1,
		Groups: map[string]*config.Group{
			"repository": {
				Paths: config.PathConfigs{{Dir: "internal/repository/**"}},
			},
			"api": {
				Paths: config.PathConfigs{{Dir: "internal/api/**"}},
				Dependencies: &config.Dependencies{
					Deny: &config.DependencyRule{
						Groups: []string{"repository"},
					},
					Allow: &config.DependencyRule{
						Patterns: []string{"internal/**"},
					},
				},
			},
		},
	}
	manager, err := groups.NewGroupManager(cfg)
	s.Require().NoError(err)

	packages := map[string]map[string]struct{}{
		"internal/api/handler": {
			"github.com/example/app/internal/repository/db": {},
		},
	}

	// when
	result, err := Check(context.Background(), "github.com/example/app", packages, manager)

	// then - deny should win even though allow pattern matches
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Violations, 1)
	assert.Equal(s.T(), "internal/repository/db", result.Violations[0].Import)
}

func (s *CheckerSuite) TestNoRules_Unrestricted() {
	// given - group with no dependency rules
	cfg := &config.Config{
		Version: 1,
		Groups: map[string]*config.Group{
			"shared": {
				Paths: config.PathConfigs{{Dir: "internal/shared/**"}},
			},
		},
	}
	manager, err := groups.NewGroupManager(cfg)
	s.Require().NoError(err)

	packages := map[string]map[string]struct{}{
		"internal/shared/utils": {
			"github.com/example/app/internal/anything": {},
		},
	}

	// when
	result, err := Check(context.Background(), "github.com/example/app", packages, manager)

	// then - no rules means unrestricted
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), result.Violations)
}

func (s *CheckerSuite) TestExternalImports_Skipped() {
	// given - strict allow rules, but external imports should be ignored
	cfg := &config.Config{
		Version: 1,
		Groups: map[string]*config.Group{
			"domain": {
				Paths: config.PathConfigs{{Dir: "internal/domain/**"}},
				Dependencies: &config.Dependencies{
					Allow: &config.DependencyRule{
						Patterns: []string{"internal/shared/**"},
					},
				},
			},
		},
	}
	manager, err := groups.NewGroupManager(cfg)
	s.Require().NoError(err)

	packages := map[string]map[string]struct{}{
		"internal/domain/user": {
			"fmt":              {},
			"context":          {},
			"github.com/other/lib": {},
		},
	}

	// when
	result, err := Check(context.Background(), "github.com/example/app", packages, manager)

	// then - stdlib and external imports are not violations
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), result.Violations)
}

func (s *CheckerSuite) TestPackageNotInAnyGroup_Skipped() {
	// given
	cfg := &config.Config{
		Version: 1,
		Groups: map[string]*config.Group{
			"domain": {
				Paths: config.PathConfigs{{Dir: "internal/domain/**"}},
				Dependencies: &config.Dependencies{
					Allow: &config.DependencyRule{
						Patterns: []string{"internal/shared/**"},
					},
				},
			},
		},
	}
	manager, err := groups.NewGroupManager(cfg)
	s.Require().NoError(err)

	packages := map[string]map[string]struct{}{
		"internal/unmatched/pkg": {
			"github.com/example/app/internal/domain/model": {},
		},
	}

	// when
	result, err := Check(context.Background(), "github.com/example/app", packages, manager)

	// then - package not in any group is not checked
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), result.Violations)
}

func (s *CheckerSuite) TestAllowOnlyMode_UnmatchedImportDenied() {
	// given - only allow specific group, import something else internal
	cfg := &config.Config{
		Version: 1,
		Groups: map[string]*config.Group{
			"shared": {
				Paths: config.PathConfigs{{Dir: "internal/shared/**"}},
			},
			"domain": {
				Paths: config.PathConfigs{{Dir: "internal/domain/**"}},
				Dependencies: &config.Dependencies{
					Allow: &config.DependencyRule{
						Groups: []string{"shared"},
					},
				},
			},
		},
	}
	manager, err := groups.NewGroupManager(cfg)
	s.Require().NoError(err)

	packages := map[string]map[string]struct{}{
		"internal/domain/user": {
			"github.com/example/app/internal/other/pkg": {},
		},
	}

	// when
	result, err := Check(context.Background(), "github.com/example/app", packages, manager)

	// then - import doesn't match any allow rule → violation
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Violations, 1)
	assert.Equal(s.T(), "internal/other/pkg", result.Violations[0].Import)
}

func (s *CheckerSuite) TestStripModulePrefix() {
	// given/when/then
	rel, ok := stripModulePrefix("github.com/example/app", "github.com/example/app/internal/domain")
	assert.True(s.T(), ok)
	assert.Equal(s.T(), "internal/domain", rel)

	_, ok = stripModulePrefix("github.com/example/app", "github.com/other/lib")
	assert.False(s.T(), ok)

	_, ok = stripModulePrefix("github.com/example/app", "fmt")
	assert.False(s.T(), ok)
}

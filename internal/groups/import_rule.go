package groups

import (
	"path"
	"strings"

	"github.com/gobwas/glob"
)

type ImportRule interface {
	Allows(fromPackage, toImport string) bool
}

type globImportRule struct {
	g glob.Glob
}

func NewGlobImportRule(pattern string) ImportRule {
	return &globImportRule{g: glob.MustCompile(pattern)}
}

func (m *globImportRule) Allows(_, toImport string) bool {
	return m.g.Match(toImport)
}

type relativeImportRule struct {
	relativePath string
}

func NewRelativeImportRule(relativePath string) ImportRule {
	return &relativeImportRule{relativePath: relativePath}
}

func (m *relativeImportRule) Allows(fromPackage, toImport string) bool {
	return path.Join(fromPackage, m.relativePath) == toImport
}

type groupImportRule struct {
	grp Group
}

func NewGroupImportRule(grp Group) ImportRule {
	return &groupImportRule{grp: grp}
}

func (m *groupImportRule) Allows(_, toImport string) bool {
	return m.grp.MatchPath(toImport)
}

type subPackageImportRule struct {
}

func NewSubPackageImportRule() ImportRule {
	return &subPackageImportRule{}
}

func (m *subPackageImportRule) Allows(fromPackage, toImport string) bool {
	return strings.HasPrefix(toImport, fromPackage+"/")
}


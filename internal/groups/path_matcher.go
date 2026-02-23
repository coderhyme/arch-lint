package groups

import "github.com/gobwas/glob"

type Matcher interface {
	Match(path string) bool
}

type globMatcher struct {
	g glob.Glob
}

func NewGlobMatcher(pattern string) Matcher {
	return &globMatcher{g: glob.MustCompile(pattern)}
}

func (m *globMatcher) Match(path string) bool {
	return m.g.Match(path)
}

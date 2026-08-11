package logs

import (
	"regexp"

	"github.com/cockroachdb/errors"
)

type Filter struct {
	includeRaw string
	include    *regexp.Regexp
	excludeRaw string
	exclude    *regexp.Regexp
}

func (f *Filter) SetInclude(pattern string) error {
	compiled, err := compilePattern(pattern)
	if err != nil {
		return errors.Wrap(err, "compile include pattern")
	}

	f.includeRaw = pattern
	f.include = compiled

	return nil
}

func (f *Filter) SetExclude(pattern string) error {
	compiled, err := compilePattern(pattern)
	if err != nil {
		return errors.Wrap(err, "compile exclude pattern")
	}

	f.excludeRaw = pattern
	f.exclude = compiled

	return nil
}

func (f *Filter) IncludePattern() string {
	return f.includeRaw
}

func (f *Filter) ExcludePattern() string {
	return f.excludeRaw
}

func (f *Filter) Matches(line string) bool {
	if f.exclude != nil && f.exclude.MatchString(line) {
		return false
	}

	if f.include != nil && !f.include.MatchString(line) {
		return false
	}

	return true
}

// compilePattern always case-folds the pattern so "error" and "ERROR" match the same lines.
func compilePattern(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, nil
	}

	compiled, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, errors.Wrap(err, "regexp compile")
	}

	return compiled, nil
}

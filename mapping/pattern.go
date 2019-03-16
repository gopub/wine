package mapping

import (
	"net/url"
	"regexp"
	"time"
)

const (
	PatternVersion   = "version"
	PatternURL       = "url"
	PatternEmail     = "email"
	PatternVariable  = "variable"
	PatternBirthDate = "birth_date"
)

var (
	versionRegexp  = regexp.MustCompile("^[1-9]\\d*(\\.\\d*)*([\\w-]+)?$")
	emailRegexp    = regexp.MustCompile("^[_a-zA-Z0-9-\\+\\.]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*(\\.[a-zA-Z]{2,3})$")
	variableRegexp = regexp.MustCompile("^[_a-zA-Z][_a-zA-Z0-9]*$")
)

type PatternMatcher interface {
	MatchPattern(pattern string, val interface{}) bool
}

type PatternMatchFunc func(pattern string, val interface{}) bool

func (f PatternMatchFunc) MatchPattern(pattern string, val interface{}) bool {
	return f(pattern, val)
}

var patternMatchers = []PatternMatcher{defaultPatternMatcher}
var patternRegexp = map[string]*regexp.Regexp{
	"version":  versionRegexp,
	"email":    emailRegexp,
	"variable": variableRegexp,
}

func PrependPatternMatcher(m PatternMatcher) {
	patternMatchers = append([]PatternMatcher{m}, patternMatchers...)
}

func AppendPatternMatcher(m PatternMatcher) {
	patternMatchers = append(patternMatchers, m)
}

func SetPatternRegexp(pattern string, reg *regexp.Regexp) {
	patternRegexp[pattern] = reg
}

func MatchPattern(pattern string, val interface{}) bool {
	for _, m := range patternMatchers {
		if m.MatchPattern(pattern, val) {
			return true
		}
	}
	return false
}

var defaultPatternMatcher PatternMatchFunc = func(pattern string, val interface{}) bool {
	if reg, ok := patternRegexp[pattern]; ok {
		if b, ok := val.([]byte); ok {
			return reg.Match(b)
		} else if s, ok := val.(string); ok {
			return reg.Match([]byte(s))
		}
	}

	switch pattern {
	case PatternURL:
		return matchURL(val)
	case PatternBirthDate:
		return matchBirthDate(val)
	}
	return false
}

func matchBirthDate(i interface{}) bool {
	s, ok := i.(string)
	if !ok {
		return false
	}

	t, e := time.Parse("2006-01-02", s)
	if e != nil {
		return false
	}

	if t.After(time.Now()) {
		return false
	}

	return true
}

func matchURL(i interface{}) bool {
	s, ok := i.(string)
	if !ok {
		return false
	}

	if u, err := url.Parse(s); err != nil {
		return false
	} else if len(u.Scheme) == 0 {
		return false
	} else if len(u.Host) == 0 {
		return false
	}

	return true
}

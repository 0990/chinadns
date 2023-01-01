package matcher

import (
	"errors"
)

type Matcher interface {
	IsMatch(domain string) bool
}

type CombineMatcher struct {
	matchers []Matcher
}

func (p *CombineMatcher) IsMatch(domain string) bool {
	for _, v := range p.matchers {
		if v.IsMatch(domain) {
			return true
		}
	}
	return false
}

func NewCombineMatcher(ms ...Matcher) Matcher {
	return &CombineMatcher{
		matchers: ms,
	}
}

func New(typ string, paths ...string) (Matcher, error) {
	var ms []Matcher
	for _, v := range paths {
		m, err := new(typ, v)
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}

	return NewCombineMatcher(ms...), nil
}

func new(typ string, file string) (Matcher, error) {
	switch typ {
	case "simple":
		return NewSimpleMatcherFromFile(file)
	case "domaintrie":
		return NewDomainTrieMatcherFromFile(file)
	case "debug":
		return NewDebugMatcher(file)
	default:
		return nil, errors.New("not support matcher type")
	}
}

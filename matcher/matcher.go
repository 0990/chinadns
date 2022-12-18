package matcher

import (
	"errors"
	"github.com/sirupsen/logrus"
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
	case "normal":
		return NewDomainMatcherFromFile(file)
	case "fast":
		return NewFastMatcherFromFile(file)
	case "test":
		return NewTestMatcher(file)
	default:
		return nil, errors.New("not support matcher type")
	}
}

type TestMatcher struct {
	*DomainMatcher
	*FastMatcher
}

func NewTestMatcher(file string) (*TestMatcher, error) {
	dm, err := NewDomainMatcherFromFile(file)
	if err != nil {
		return nil, err
	}

	fm, err := NewFastMatcherFromFile(file)
	if err != nil {
		return nil, err
	}

	return &TestMatcher{
		DomainMatcher: dm,
		FastMatcher:   fm,
	}, nil
}

func (p *TestMatcher) IsMatch(domain string) bool {
	a := p.DomainMatcher.IsMatch(domain)
	b := p.FastMatcher.IsMatch(domain)

	if a != b {
		logrus.WithField("domain", domain).Warnf("domainMatch:%v fastMatch:%v", a, b)
	}
	return a
}

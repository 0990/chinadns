package matcher

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Matcher interface {
	IsMatch(domain string) bool
}

func New(typ string, file string) (Matcher, error) {
	switch typ {
	case "normal":
		return NewDomainMatcherFromFile(file)
	case "fast":
		return NewFastMatcherFromFile(file)
	case "combine":
		return NewComineMatcher(file)
	default:
		return nil, errors.New("not support matcher type")
	}
}

type ComineMatcher struct {
	*DomainMatcher
	*FastMatcher
}

func NewComineMatcher(file string) (*ComineMatcher, error) {
	dm, err := NewDomainMatcherFromFile(file)
	if err != nil {
		return nil, err
	}

	fm, err := NewFastMatcherFromFile(file)
	if err != nil {
		return nil, err
	}

	return &ComineMatcher{
		DomainMatcher: dm,
		FastMatcher:   fm,
	}, nil
}

func (p *ComineMatcher) IsMatch(domain string) bool {
	a := p.DomainMatcher.IsMatch(domain)
	b := p.FastMatcher.IsMatch(domain)

	if a != b {
		logrus.WithField("domain", domain).Errorf("domainMatch:%v fastMatch:%v", a, b)
	}
	return a
}

package matcher

import "github.com/sirupsen/logrus"

// 所有匹配器一起用，看哪个效果好
type DebugMatcher struct {
	*SimpleMatcher
	*DomainTrieMatcher
	fileName string
}

func NewDebugMatcher(file string) (*DebugMatcher, error) {
	dm, err := NewSimpleMatcherFromFile(file)
	if err != nil {
		return nil, err
	}

	fm, err := NewDomainTrieMatcherFromFile(file)
	if err != nil {
		return nil, err
	}

	return &DebugMatcher{
		SimpleMatcher:     dm,
		DomainTrieMatcher: fm,
		fileName:          file,
	}, nil
}

func (p *DebugMatcher) IsMatch(domain string) bool {
	a := p.SimpleMatcher.IsMatch(domain)
	b := p.DomainTrieMatcher.IsMatch(domain)

	if a != b {
		logrus.WithFields(logrus.Fields{
			"domain": domain,
			"file":   p.fileName,
		}).Warnf("simpleMatch:%v domainTrieMatch:%v", a, b)
	}
	return b
}

package matcher

import (
	"bufio"
	"os"
	"strings"
)

type DomainMatcher struct {
	domains []string
}

func NewDomainMatcherFromFile(file string) (*DomainMatcher, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dm := &DomainMatcher{}

	reader := bufio.NewReader(f)
	for {
		var line []byte
		data, isPrefix, err := reader.ReadLine()
		if nil != err {
			break
		}
		line = append(line, data...)
		if isPrefix {
			continue
		}
		dm.addRule(string(line))
	}
	return dm, nil
}

func (p *DomainMatcher) addRule(rule string) {
	rule = strings.TrimSpace(rule)

	p.domains = append(p.domains, rule)
}

func (p *DomainMatcher) IsMatch(domain string) bool {
	for _, dr := range p.domains {
		if strings.HasSuffix(domain, dr) {
			if domain == dr || strings.HasSuffix(domain, "."+dr) {
				return true
			}
		}
	}
	return false
}

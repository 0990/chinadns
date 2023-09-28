package matcher

import (
	"bufio"
	"os"
)

type DomainTrieMatcher struct {
	tr *domainTrie
}

func NewDomainTrieMatcherFromFile(file string) (*DomainTrieMatcher, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tr := &domainTrie{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		tr.Add(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &DomainTrieMatcher{tr: tr}, nil
}

func (m *DomainTrieMatcher) IsMatch(domain string) bool {
	return m.tr.Contain(domain)
}

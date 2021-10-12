package matcher

import (
	"bufio"
	"os"
	"strings"
)

type FastMatcher struct {
	domains map[string]struct{}
}

func NewFastMatcherFromFile(file string) (*FastMatcher, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dm := &FastMatcher{
		domains: map[string]struct{}{},
	}

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

func (p *FastMatcher) addRule(rule string) {
	rule = strings.TrimSpace(rule)

	p.domains[rule] = struct{}{}
}

func (r *FastMatcher) IsMatch(domain string) bool {
	combines := r.combineDomains(domain)
	for _, v := range combines {
		if _, exist := r.domains[v]; exist {
			return true
		}
	}

	return false
}

func (r *FastMatcher) combineDomains(domain string) []string {
	ss := strings.Split(domain, ".")

	length := len(ss)

	var ret []string
	for i := 0; i <= length-2; i++ {
		d := strings.Join(ss[i:], ".")
		ret = append(ret, d)
	}

	return ret
}

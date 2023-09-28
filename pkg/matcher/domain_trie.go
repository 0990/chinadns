package matcher

import "strings"

type domainTrie struct {
	children map[string]*domainTrie
	end      bool
}

func (tr *domainTrie) Add(domain string) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return
	}

	domain = strings.Trim(domain, ".")
	if domain == "" {
		tr.end = true
		tr.children = nil
		return
	}

	node := tr

	labels := strings.Split(domain, ".")

	for i := len(labels) - 1; i >= 0; i-- {
		if node.end {
			return
		}

		if node.children == nil {
			node.children = make(map[string]*domainTrie)
		}
		label := labels[i]
		if node.children[label] == nil {
			node.children[label] = &domainTrie{}
		}
		node = node.children[label]
	}
	node.end = true
}

func (tr *domainTrie) Contain(domain string) bool {
	if tr == nil {
		return false
	}

	domain = strings.Trim(domain, ".")
	labels := strings.Split(domain, ".")

	node := tr
	if node.end {
		return true
	}

	for i := len(labels) - 1; i >= 0; i-- {
		label := labels[i]
		node = node.children[label]
		if node == nil {
			return false
		}

		if node.end {
			return true
		}
	}

	return false
}

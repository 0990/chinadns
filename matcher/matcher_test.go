package matcher

import (
	"bufio"
	"os"
	"testing"
)

func Test_Matcher(t *testing.T) {
	m, err := New("normal", "gfwlist.txt")
	if err != nil {
		t.Fatal(err)
	}

	tbls := []struct {
		domain  string
		isMatch bool
	}{
		{"www.google.com", true},
		{"www.baidu.com", false},
	}

	for _, v := range tbls {
		ret := m.IsMatch(v.domain)
		if ret != v.isMatch {
			t.Errorf("domain:%s ret:%v expect:%v", v.domain, ret, v.isMatch)
		}
	}
}

func TestMatcherAllRule(t *testing.T) {
	file := "gfwlist.txt"
	m, err := New("normal", file)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

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

		domain := string(line)
		if !m.IsMatch(domain) {
			t.Errorf("%v not match", domain)
		}
	}
}

func BenchmarkDomainMatcher_IsMatch(b *testing.B) {
	m, err := New("normal", "gfwlist.txt")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		m.IsMatch("baidu.com")
	}
}

func TestDomainTrie_Contain(t *testing.T) {
	tr := &domainTrie{}

	tr.Add("www.google.com")
	tr.Add(".google.com")

	contain := tr.Contain("cn.google.com")

	if contain == false {
		t.Fail()
	}
}

func Test_Matcher_Compare(t *testing.T) {
	m, err := New("test", "chnlist.txt")
	if err != nil {
		t.Fatal(err)
	}

	m.IsMatch("access.open.uc.cn")
}

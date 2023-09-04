package chinadns

import "sync"

type MultiError struct {
	errs []error
	sync.Mutex
}

func (m *MultiError) Add(err error) {
	m.Lock()
	defer m.Unlock()
	m.errs = append(m.errs, err)
}

func (m *MultiError) Error() string {
	m.Lock()
	defer m.Unlock()
	if len(m.errs) == 0 {
		return ""
	}

	var s string
	for _, v := range m.errs {
		s += v.Error()
		s += ";"
	}

	s = s[:len(s)-1]
	return s
}

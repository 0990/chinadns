package chinadns

type MultiError struct {
	errs []error
}

func (m *MultiError) Add(err error) {
	m.errs = append(m.errs, err)
}

func (m *MultiError) Error() string {
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

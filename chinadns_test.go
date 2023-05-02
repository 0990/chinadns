package chinadns

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"
)

func Test_CreateConfig(t *testing.T) {

	cfg := Config{
		Listen:      "0.0.0.0:53",
		UDPMaxBytes: 4096,
		Timeout:     2,
		DNSChina:    []string{"114.114.114.114"},
		DNSAbroad:   []string{"8.8.8.8"},
		LogLevel:    "debug",
		Domain2IP: map[string]string{
			"a.b": "127.0.0.1",
		},
	}

	c, _ := json.MarshalIndent(cfg, "", "   ")
	ioutil.WriteFile("chinadns.json", c, 0644)
}

func Test_MultiError(t *testing.T) {
	errs := &MultiError{}
	errs.Add(errors.New("hello"))
	errs.Add(errors.New("world"))
	if errs.Error() != "hello;world" {
		t.Failed()
	}
}

package chinadns

import (
	"encoding/json"
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
	}

	c, _ := json.MarshalIndent(cfg, "", "   ")
	ioutil.WriteFile("chinadns.json", c, 0644)
}

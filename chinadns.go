package chinadns

type Config struct {
	Listen      string `json:"listen"`
	UDPMaxBytes int    `json:"udp-max-bytes"`
	Timeout     int    `json:"timeout"` //查询超时时间

	GFWPath   string   `json:"gfw_path"`
	DNSChina  []string `json:"dns-china"`  //国内dns
	DNSAbroad []string `json:"dns-abroad"` //海外dns,可信dns

	LogLevel string `json:"log_level"`
}

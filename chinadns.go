package chinadns

type Config struct {
	Listen         string `json:"listen"`
	UDPMaxBytes    int    `json:"udp-max-bytes"`
	Timeout        int    `json:"timeout"`          //查询超时时间
	CacheExpireSec int    `json:"cache_expire_sec"` //缓存超时时间

	GFWPath   []string `json:"gfw_path"`
	DNSChina  []string `json:"dns-china"`  //国内dns
	DNSAbroad []string `json:"dns-abroad"` //海外dns,可信dns

	ChnPath string `json:"chn_path"` //国内ip列表

	LogLevel string `json:"log_level"`
}

package chinadns

type Config struct {
	Listen         string `json:"listen"`
	UDPMaxBytes    int    `json:"udp-max-bytes"`
	Timeout        int    `json:"timeout"`          //查询超时时间
	CacheExpireSec int    `json:"cache_expire_sec"` //缓存超时时间

	//GFWPath   []string `json:"gfw_path"`
	DNSChina  []string `json:"dns-china"`  //国内dns
	DNSAbroad []string `json:"dns-abroad"` //海外dns,可信dns

	ChnIP     []string `json:"chn_ip"` //国内ip列表
	ChnDomain string   `json:"chn_domain"`
	GfwDomain string   `json:"gfw_domain"`

	LogLevel  string `json:"log_level"`
	PProfPort int    `json:"pprof_port"`
}

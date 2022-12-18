package chinadns

type Config struct {
	Listen         string `json:"listen"`
	UDPMaxBytes    int    `json:"udp-max-bytes"`
	Timeout        int    `json:"timeout"`          //查询超时时间
	CacheExpireSec int    `json:"cache_expire_sec"` //缓存超时时间

	Domain2IP map[string]string `json:"domain2ip"` //自定义dns,优先于domain2attr

	DNSChina      []string `json:"dns-china"`       //国内dns
	DNSAbroad     []string `json:"dns-abroad"`      //海外dns,可信dns
	DNSAbroadAttr string   `json:"dns-abroad-attr"` //海外dns特性 noipv4 noipv6 nocname

	ChnIP     []string `json:"chn_ip"` //国内ip列表
	ChnDomain string   `json:"chn_domain"`
	GfwDomain string   `json:"gfw_domain"`

	LogLevel  string `json:"log_level"`
	PProfPort int    `json:"pprof_port"`
}

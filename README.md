# chinadns
防dns污染的域名解析服务,需配合透明代理使用

## 特性
* 手动指定国内，国外上游域名服务器，上游域名请求支持udp,tcp,dns over http
* 支持域名缓存
* 支持自定义域名解析
* 支持海外dns屏蔽ipv4或ipv6解析

## 配置
```
{
	"listen": "0.0.0.0:53",
	"udp-max-bytes": 4096,
	"timeout": 5,
	"cache_expire_sec": 600,
	"domain2ip": {},
	"dns-china": [
		"114.114.114.114"
	],
	"dns-abroad": [
		"8.8.8.8"
	],
	"chn_ip": [
		"chnroute.txt",
		"chnroute6.txt"
	],
	"chn_domain": [
		"chinalist.txt"
	],
	"gfw_domain": [
		"gfwlist.txt"
	],
	"log_level": "error"
}
```

## 原理
当有dns请求到来时，将按以下顺序处理</br>
1,匹配chn_domain文件中的域名，匹配到则使用dns-china来解析</br>
2,匹配gfw_domain文件中的域名，匹配到则使用dns-abroad来解析</br>
3,以上两种都匹配不到，则同时使用dns-china和dns-abroad解析</br>
当dns-china解析ip为国外时，取dns-abroad解析结果，否则取dns-china解析结果


### 配置详解
#### cache_expire_sec
dns缓存时间（秒），当<=0代表不启用dns缓存
#### domain2ip
自定义域名解析（支持ipv6)，格式为 "域名":"ip1;ip2;ip3",当ip配置为<br>
0:0:0:0,代表禁用ipv4解析<br>
0:0:0:0:0:0:0:0或者::,代表禁用ipv6解析<br>
禁用后，返回的answer域为空
#### dns-china dns-abroad
国内外上游dns服务器，格式为protocol@ip:port,可省略为ip<br>
protocol支持udp,tcp,doh(dns over http)
#### chn_ip
国内ip列表文件，用于原理步骤3中判定是否为国外ip所用

## Thanks
[cherrot/gochinadns](https://github.com/cherrot/gochinadns)  


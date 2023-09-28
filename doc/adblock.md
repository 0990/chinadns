# 广告过滤
chinadns支持广告过滤，需要有广告过滤dns服务器，比如AdGuardHome

## AdGuardHome过滤广告原理
当AdGuardHome识别到广告域名时，会返回
* ipv4情况下返回0.0.0.0
* ipv6情况下返回::
* https情况下返回fake-for-negative-caching.adguard.com

利用这个特性，可以实现chinadns的广告过滤

## 流程
1. chinadns会将请求并发请求到dns-adblock中
2. 当此域名被比AdGuardHome设别为广告域名时，会返回上述结果
3. 当chinadns收到上述结果时，则认为此域名为广告域名，直接返回结果，否则使用dns-china或者dns-abroad解析

### 配置
```json
{
  "dns-adblock": [],
  "dns-adblock-reply": []
}
```
#### dns-adblock
广告过滤dns服务器
#### dns-adblock-reply
广告过滤dns服务器识别域名为广告时的dns解析回复结果，当判断为是广告域名时，会直接返回结果给请求方<br>
默认值为 "0.0.0.0", "::", "fake-for-negative-caching.adguard.com"



# Mag

## About

在不得已的情况下，对多个资产进行目录爆破会带来几个问题：

1. 服务器负载过大。
2. 效率较低。
3. 返回大量重复且不易查阅结果

综合各种情况，目录扫描大多数时候都吃力不讨好，所以大多目录扫描工具更偏向于针对单个目标，即使有多目标功能也并未对此做响应优化。

针对单目标的话，只需要简单的根据情况加上延迟即可。

但如果是多目标，就有很大的操作空间。

假设100个目标，1000个路径，按照一般扫描器，一个目标扫完1k个路径，再进行下一个目标的扫描，这样十分容易造成服务器崩溃。

如果加个延迟呢？那效率又降低了。当然聪明的开发者肯定不会死等，会在延迟期间调度扫描至其他目标。但如此还是局限于前面的几个目标。

针对这个问题，有个十分简单且有效的办法，就是基于路径。选出路径A，扫目标1、目标2、目标3....目标100；然后再选出路径B重复扫描。这样将轮询扫描的时间都当做给服务器的缓冲，在一定(哪怕不是很多)程度上降低了负载且兼顾了效率。

正好想学习下Golang开发，又无意间发现这个想法在[meg](https://github.com/tomnomnom/meg)中得到实现，可惜该项目停更许久。抱着学习的目的，在该项目的基础上诞生了[mag](https://github.com/KagamigawaMeguri/mag)，其中也借鉴了不少其他项目。

当然，该项目不仅限于此，还会加入更多点子。哪怕没用，只为好玩。毕竟别人的工具用久了，总会想写点属于自己的东西。

### Improve

基于个人需求对原项目[meg](https://github.com/tomnomnom/meg)的一些改动

- 重构HTTP请求封装

  主要为了练手，参考了几个项目的封装代码，尝试性的重构，希望能写的更优雅一点，以便后续开发。

- 加入简单的返回包过滤

  主要用于去重，过滤重复页面。此处参考了[crawlergo](https://github.com/Qianlitp/crawlergo)，对返回包进行MD5计算过滤完全相同界面；又加入返回包长度检测用于过滤大部分相同极少部分不同的页面。

- 增加Simhash算法去重

- 部分代码采用更优写法

  此处参考了网上一些优化文章，尝试性的对代码进行优化，以求提高性能。部分代码bechmark测试通过。

- 修改命令行参数

  因为个人很喜欢[httpx](https://github.com/projectdiscovery/httpx)，所以将命令行参数向该项目靠拢，便于记忆。

- 增加TLS1.0和TLS1.1支持

- 扫指定目录和扫备份功能细分。

  在大多数情况下，扫备份文件足矣，但

- 增加数据库缓存

### Todo

- 采用更优的页面相似度算法

  1.目前翻阅了一些有关论文，准备着手写个基于DOM的算法

  2.参考SQLMAP的页面相似度算法写个Go版的

  3.使用SimHash算法：后续可考虑增加默认指纹用于区分常见404/报错页面/WAF页面

- 进度显示

- 备份扫描

- GUI或WebUI

- REST API

  用于接入到其他程序

- 动态调整延迟

  简略方案是根据延迟每轮进行调整。更优方案是使用深度学习进行动态调优。

- 分布式

  视情况而定，可能会考虑接入其他分布式工具

- 支持HTTP/2与动态调整持久连接相关参数

  目前请求基本都是HTTP1.1且关闭持久连接，后续会根据返回情况进行动态调整以求降低负载。可能会用到深度学习进行研判。

## Usage

```
兼顾效率与负载的多任务目录扫描器

Usage:
  mag.exe [flags]

Flags:
INPUT:
   -l, -list string  目标主机文件 (default "./host.txt")
   -w, -path string  路径字典文件 (default "./path.txt")

OUTPUT:
   -o, -output string   输出路径 (default "./out")
   -do, -disableoutput  禁用输出

CONFIGURATIONS:
   -x, -method string            自定义请求方法
   -body string                  自定义请求包
   -proxy, -http-proxy string  设置代理 (eg http://127.0.0.1:8080)
   -H, -header string[]          自定义请求头
   -d, -delay duration           扫描时相同host间最小延迟 (eg: 200ms, 1s) (default 50ms)
   -timeout int                  请求超时时间 (default 10)
   -f, -follow                   是否允许重定向 (default true)
   -slow                         服务器极度友好模式
   -t, -thread int               最大线程数 (default 25)
   -random-agent                 是否启动随机UA-待开发 (default true)

MATCHERS:
   -mc, -match-code string    匹配指定状态码 (eg: -mc 200,302)
   -ml, -match-length string  匹配指定长度 (eg: -ml 100,102)
   -ms, -match-string string  匹配指定字符串 (eg: -ms admin)
   -mr, -match-regex string   匹配指定正则 (eg: -mr admin)

FILTERS:
   -fc, -filter-code string    过滤指定状态码 (eg: -fc 403,401)
   -fl, -filter-length string  过滤指定长度 (eg: -ml 100,102)
   -fs, -filter-string string  过滤指定长度 (eg: -fs admin)
   -fr, -filter-regex string   过滤指定正则 (eg: -fe admin)

DEBUG:
   -v, -verbose  verbose mode
```

## Acknowledgement

https://github.com/tomnomnom/meg

https://github.com/projectdiscovery/httpx

https://github.com/OJ/gobuster

https://github.com/Qianlitp/crawlergo
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

- 部分代码采用更优写法

  此处参考了网上一些优化文章，尝试性的对代码进行优化，以求提高性能。部分代码bechmark测试通过。

- 修改命令行参数

  因为个人很喜欢[httpx](https://github.com/projectdiscovery/httpx)，所以将命令行参数向该项目靠拢，便于记忆。

- 增加TLS1.0和TLS1.1支持

- 其他修改

  杂七杂八的修改，具体见代码

### Todo

- 采用更优的页面相似度算法

  1.目前翻阅了一些有关论文，准备着手写个基于DOM的算法

  2.参考SQLMAP的页面相似度算法写个Go版的

  3.使用SimHash算法

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
兼顾效率与负载的多任务服务器友好型目录扫描器

用法:
  mag [pathsFile] [hostsFile] [outputDir]

请求:
  -X, -method <string>                  设置请求方法(默认GET)
  -H, -header <string>                  设置请求头
  -b, -body <string>                    设置POST请求体
  -t, -threads <int>                    设置并发数(默认20)
  -d, -delay <int>                      设置相同host间的延迟(默认5000ms)
  -timeout <int>                        设置超时时间(默认10000ms)
  -proxy <string>                       设置代理
  -fr, -follow-redirects                允许重定向
  -no-headers                           不设置请求头
  -slow                                 服务器极度友好模式

匹配:
  -ms, -match-string <string>           检测到指定字符串则保存
  -mr, -match-regex <string>            检测到指定regex则保存
  -mc, -match-code <int>                检测到指定状态码则保存：-match-code 200,301

过滤:
  -fs, -filter-string <string>          检测到指定字符串则跳过
  -fe, -filter-regex <string>           检测到指定regex则跳过
  -fl, -filter-length <int>             检测到指定长度则跳过：-match-code 200,301

DEBUG:
  -v,  -verbose                         Verbose mode

默认路径:
  pathsFile: ./paths
  hostsFile: ./hosts
  outputDir: ./out
```

## Acknowledgement

https://github.com/tomnomnom/meg

https://github.com/projectdiscovery/httpx

https://github.com/OJ/gobuster

https://github.com/Qianlitp/crawlergo
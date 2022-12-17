# 我的世界加速ip

## 如何运行
1. [下载release](https://github.com/sduoduo233/go-mcproxy/releases/latest)
2. 执行 `./mcproxy`

## 参数说明
```
Usage of ./mcproxy:
  -description string
        server description
  -fakeping
        fake ping
  -favicon string
        server icon (default "favicon.png")
  -help
        print help message
  -listen string
        local listening address (default "127.0.0.1:25565")
  -max int
        max player (default 20)
  -remote string
        remote forward address (default "mc.hypixel.net:25565")
```
- description: 服务器的描述，不指定代表使用原服务器的
- fakeping：假延迟
- listen: 本地监听的端口
- remote：原服务器的地址
- favicon: 服务器图标，必须是png格式，大小64x64
 
## 用户认证
相关函数在`auth.go`中，默认允许所有人加入
```
func allowJoin(username string) (bool, string) {
	return true, "You are not whitelisted."
}
```
第一个返回值代表是否允许用户加入，第二个是不允许加入的原因
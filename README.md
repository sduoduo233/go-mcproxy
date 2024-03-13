# 我的世界加速ip

## 如何运行
1. [下载release](https://github.com/sduoduo233/go-mcproxy/releases/latest)
2. 执行 `./mcproxy`

## 参数说明
```
Usage of ./mcproxy:
  -config string
        path to config.json (default "config.json")
```

`-config` 配置文件路径

## 配置文件说明
 
```json

{
    "listen": "0.0.0.0:25565",
    "description": "hello\nworld",
    "remote": "mc.hypixel.net:25565",
    "max_player": 20,
    "ping_mode": "fake",
    "fake_ping": 0,
    "rewrite_host": "mc.hypixel.net",
    "rewrite_port": 25565,
    "auth": "none",
    "whitelist": [
        "L1quidBounce"
    ],
    "blacklist": [

    ]
}

```

`listen`: 服务器监听地址

`description`: MOTD

`remote`: 反向代理的源服务器

`max_player`: 最大玩家

`ping_mode`: 相应 ping 的方法，可以是 `real`（真实延迟），或 `fake`（假延迟）

`rewrite_host`：修改客户端发送的服务器地址（可以用来绕过 Hypixel 的地址检测）

`rewrite_port`：修改客户端发送的服务器端口

`auth`：用户名认证，可以是 `none`, `blacklist` 或 `whitelist`
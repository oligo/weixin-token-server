# 微信公众号access token server

## Features

1. 支持多个公众号的token管理
2. 可以持久化到外部存储（当前仅支持磁盘文件)
3. 可以自动更新access token

## APIs


* curl -XGET 'http://127.0.0.1:8080/api/v1/token?appId=weixin-app-id'

Response:
```json
  {
    "access_token": "25_leRR-weixin-access-token",
    "expires_in":7130
  }
```

Key `access_token`为微信公众号的access token, `expires_in`是token的过期时间，在过期前access token会自动刷新。
客户端可以根据`expires_in`来设置缓存过期时间，或下次重新获取token的时间.


## Run

token-server的最新docker image已经push到Docker Hub, 搜索oligo/token-server即可找到。运行方式：

```shell
  docker run -d -v "${HOME}/.token-server:/root/.token-server" -p '127.0.0.1:8080:8080' oligo/token-server
```

## Configuration

token-server有一个简单的yaml配置文件，可以放置在`$HOME/.token-server/`目录下，也可以放置在进程同一目录下面。配置key如下：

```yaml
 # weixin mp credential:
  credentials:
    - appId: wx-appid-example
      appSecret: wx-app-secret-example
    
    - appId: wx-appid-example2
      appSecret: wx-app-secret-example2

  # how often to check expiration time of access token
  check:
    interval: 10s
```

 token-server在停止时，会保存获取到的access token到`$HOME/.token-server/token.json`，启动时，也会尝试从这个文件加载以前的数据，以防止反复的请求微信API。



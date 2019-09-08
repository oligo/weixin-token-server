# 微信公众号access token server

在公众号开发时，经常需要管理微信Oauth2签发的access token, 这个项目提供的就是这样一个小的服务，开箱（容器）即用，非常适合微服务场景下部署。当前仅支持持久化到磁盘文件，未来可以考虑支持redis等外部存储。现阶段不支持HA，适合于中小项目或个人项目使用。

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

Key `access_token`为微信公众号的access token, `expires_in`是token的过期时间(TTL)，在过期前access token会自动刷新。
客户端可以根据`expires_in`来设置缓存过期时间，或下次重新获取token的时间.


## Run

token-server的最新docker image已经push到[Docker Hub](https://hub.docker.com/r/oligo/token-server), 搜索oligo/token-server即可找到。运行方式：

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

 ## HA

 如前面所讲到的，token-server在内部维护了token的状态，多节点部署时token的刷新不会同步，因此目前来说还不适合高可用的部署方式。推荐使用时使用一些进程管理器去管理token-server的生命周期，如果在容器环境下，如docker swarm/k8s，设置服务/容器失败重启就可以了。另外，access token的消费方也建议根据`expires_in`来合理的缓存之，这样可以避免频繁请求token-server, 也能应对部分token-server挂掉的场景（只有token-server能很快恢复）。



# 微信公众号access token server

## APIs


* curl -XGET 'http://127.0.0.1:8080/api/v1/token?appId=weixin-app-id'

Response:

  {
    "access_token": "25_leRR-weixin-access-token",
    "expires_in":7130
  }
  
Key `access_token`为微信公众号的access token, `expires_in`是token的过期时间，在过期前access token会自动刷新。


## Run

token-server的最新docker image已经push到Docker Hub, 搜索oligo/token-server即可找到。运行方式：

  docker run --rm -v "`pwd`/config.yml:/config.yml" oligo/token-server
  
## Configuration

token-server有一个简单的yaml配置文件，可以放置在`$HOME/.token-server/`目录下，也可以放置在进程同一目录下面。配置key如下：
  
  # weixin mp credential:
  credential:
    appId: wx-appid-example
    appSecret: wx-app-secret-example

  # how often to check expiration time of access token
  check:
    interval: 10s
 
 token-server在停止时，会保存获取到的access token到`$HOME/.token-server/token.json`，启动时，也会尝试从这个文件加载以前的数据，以防止反复的请求微信API。



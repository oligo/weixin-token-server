package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type wechatCredential struct {
	appId     string
	appSecret string
}

type WechatAccessToken struct {
	AccessToken string    `json:"accessToken"`
	ExpireTime  int       `json:"expireTime"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ticker      *time.Ticker
	mutex       *sync.Mutex
	cred        *wechatCredential
	stopChan    chan struct{}
	db          io.ReadWriteCloser
	*json.Encoder
	*json.Decoder
}

func newAccessToken(cred *wechatCredential, interval time.Duration) *WechatAccessToken {
	t := &WechatAccessToken{
		ticker:   time.NewTicker(interval),
		cred:     cred,
		stopChan: make(chan struct{}, 1),
		mutex:    &sync.Mutex{},
	}

	diskFile, err := os.OpenFile(path.Join(appHome, "token.json"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		log.Printf("open token.json failed: %v\n", err)
	}

	t.db = diskFile
	t.Encoder = json.NewEncoder(diskFile)
	t.Decoder = json.NewDecoder(diskFile)

	return t
}

func (t *WechatAccessToken) ExpiresIn() time.Duration {
	return (time.Duration(t.ExpireTime*1000*1000*1000) - time.Since(t.UpdatedAt)).Round(time.Second)
}

func (t *WechatAccessToken) Tick() {
	t.Decode(t)
	t.Update()

	for {
		select {
		case <-t.ticker.C:
			t.mutex.Lock()
			if t.ExpiresIn() < 5*60*time.Second {
				t.queryForToken()
				log.Println("access token refreshed")
			}
			t.mutex.Unlock()
		case <-t.stopChan:
			t.ticker.Stop()
			log.Println("quit access token ticker")
			return
		}
	}

}

// Update force the update of wechat access token
func (t *WechatAccessToken) Update() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.UpdatedAt.IsZero() || t.ExpiresIn() < 5*60*time.Second {
		t.queryForToken()
		log.Println("access token updated")
	} else {
		log.Println("no need to update")
	}
}

// Close shutdown the tick cycles
func (t *WechatAccessToken) Close() {
	t.stopChan <- struct{}{}
	t.Encode(t)
	t.db.Close()
}

func (t *WechatAccessToken) queryForToken() error {
	url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + t.cred.appId + "&secret=" + t.cred.appSecret

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	var respMap map[string]interface{}
	json.Unmarshal(bodyData, &respMap)

	// 读取access_token
	if _, ok := respMap["access_token"]; !ok {
		log.Fatalf("Failed to update access token, errcode: %v, errmsg: %s\n", respMap["errcode"], respMap["errmsg"].(string))
		return nil
	}

	t.AccessToken = respMap["access_token"].(string)
	t.ExpireTime = int(respMap["expires_in"].(float64))
	t.UpdatedAt = time.Now()

	return nil
}

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type wechatCredential struct {
	AppID     string `mapstructure:"appId"`
	AppSecret string `mapstructure:"appSecret"`
}

type AccessToken struct {
	AppID      string    `json:"appId"`
	Token      string    `json:"accessToken"`
	ExpireTime int       `json:"expireTime"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type AccessTokenHolder struct {
	*AccessToken
	ticker   *time.Ticker
	mutex    *sync.Mutex
	cred     *wechatCredential
	stopChan chan struct{}
}

func newAccessTokenHolder(cred *wechatCredential, interval time.Duration) *AccessTokenHolder {
	t := &AccessTokenHolder{
		ticker:   time.NewTicker(interval),
		cred:     cred,
		stopChan: make(chan struct{}, 1),
		mutex:    &sync.Mutex{},
	}

	return t
}

func (t *AccessTokenHolder) ExpiresIn() time.Duration {
	if t.UpdatedAt.IsZero() {
		return 0
	}

	return (time.Duration(t.ExpireTime*1000*1000*1000) - time.Since(t.UpdatedAt)).Round(time.Second)
}

func (t *AccessTokenHolder) Tick() {
	t.Update()

	for {
		select {
		case <-t.ticker.C:
			t.mutex.Lock()
			if t.ExpiresIn() < 3*60*time.Second {
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
func (t *AccessTokenHolder) Update() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.AccessToken == nil || t.ExpiresIn() < 3*60*time.Second {
		t.queryForToken()
		log.Println("access token updated")
	} else {
		log.Println("no need to update")
	}
}

// Close shutdown the tick cycles
func (t *AccessTokenHolder) Close() {
	t.stopChan <- struct{}{}
}

func (t *AccessTokenHolder) queryForToken() error {
	url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + t.cred.AppID + "&secret=" + t.cred.AppSecret

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
		log.Printf("failed to update access token, errcode: %v, errmsg: %s\n", respMap["errcode"], respMap["errmsg"].(string))
		t.AccessToken = &AccessToken{}
		return nil
	}

	t.AccessToken = &AccessToken{
		AppID:      t.cred.AppID,
		Token:      respMap["access_token"].(string),
		ExpireTime: int(respMap["expires_in"].(float64)),
		UpdatedAt:  time.Now(),
	}

	return nil
}

// AccessTokenPool manages all access token
type AccessTokenPool struct {
	pool  map[string]*AccessTokenHolder
	store TokenStore
}

func NewAccessTokenPool(store TokenStore) *AccessTokenPool {
	return &AccessTokenPool{
		pool:  make(map[string]*AccessTokenHolder),
		store: store,
	}
}

// Put puts a new access token holder in the pool
func (p *AccessTokenPool) Put(holder *AccessTokenHolder) error {
	if holder.cred.AppID == "" {
		return errors.New("no appId in credential")
	}

	// load previous saved token from token store
	token, err := p.store.Load(holder.cred.AppID)
	if err != nil {
		log.Printf("loading token from token store failed: %v", err)
	}

	holder.AccessToken = token

	p.pool[holder.cred.AppID] = holder

	return nil
}

// Get gets a added access token holder from the pool
func (p *AccessTokenPool) Get(appID string) (*AccessTokenHolder, error) {
	holder, ok := p.pool[appID]
	if !ok {
		return nil, errors.New("access token holder not found for " + appID)
	}

	return holder, nil
}

func (p *AccessTokenPool) SaveAll() error {
	tokens := make([]AccessToken, 0)
	for _, holder := range p.pool {
		tokens = append(tokens, *holder.AccessToken)
	}

	p.store.Save(tokens...)

	return nil
}

func (p *AccessTokenPool) Close() {
	p.SaveAll()
	for _, holder := range p.pool {
		holder.Close()
	}
}

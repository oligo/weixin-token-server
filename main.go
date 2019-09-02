package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

var (
	appHome     string
	accessToken *WechatAccessToken
	signalChan  chan os.Signal
)

func init() {
	signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	initHome()
	loadConfig()
}

func main() {

	cred := &wechatCredential{
		appId:     viper.GetString("credential.appId"),
		appSecret: viper.GetString("credential.appSecret"),
	}

	accessToken = newAccessToken(cred, viper.GetDuration("check.interval"))

	go func() {
		accessToken.Tick()
	}()

	server := newServer(":8080")

	go func() {
		log.Println("Listening on http://0.0.0.0:8080")

		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	hook := Hook{sigChan: signalChan}

	hook.register(func() {
		accessToken.Close()
	})

	hook.register(func() {
		log.Println("Shutting down the access token server")

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		server.Shutdown(ctx)
	})

	hook.listen()

}

func loadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.token-server")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}

func initHome() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	appHome = path.Join(home, ".token-server")

	_, err = os.Stat(appHome)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(appHome, 0700)
		if err != nil {
			panic(err)
		}
	}
}

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

const diskStoreFile = "token.json"

var (
	appHome    string
	cred       *wechatCredential
	pool       *AccessTokenPool
	signalChan chan os.Signal
)

func init() {
	signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	initHome()
	loadConfig()
	pool = NewAccessTokenPool(NewDiskStore(path.Join(appHome, diskStoreFile)))
}

func main() {

	var credentials []wechatCredential

	err := viper.UnmarshalKey("credentials", &credentials)
	if err != nil {
		log.Fatalln("load config failed")
	}

	log.Println(credentials)

	for _, cred := range credentials {
		log.Printf("loading %s\n", cred.AppID)
		holder := newAccessTokenHolder(&cred, viper.GetDuration("check.interval"))
		err = pool.Put(holder)
		if err != nil {
			panic(err)
		}

		go func() {
			holder.Tick()
		}()
	}

	server := newServer(":8080")

	go func() {
		log.Println("Listening on http://0.0.0.0:8080")

		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	hook := Hook{sigChan: signalChan}

	hook.register(func() {
		log.Println("Shutting down the access token server")

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		server.Shutdown(ctx)
	})

	hook.register(func() {
		pool.Close()
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

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
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

	for _, cred := range credentials {
		log.Printf("loading %s\n", cred.AppID)
		holder := newAccessTokenHolder(cred, viper.GetDuration("check.interval"))
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
			log.Println(err)
		}
	}()

	<-signalChan
	log.Println("process interrupted...")
	log.Println("shutting down the access token server")
	pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = server.Shutdown(ctx)
	if err != nil {
		log.Println(err)
	}
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(appHome)
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

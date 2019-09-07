package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type APIServerMux struct {
	logger *log.Logger
	mux    *http.ServeMux
}

func (s *APIServerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.logger.Printf("remote: %s, UA: %s, method: %s, path: %s", r.RemoteAddr, r.UserAgent(), r.Method, r.RequestURI)
	s.mux.ServeHTTP(w, r)
}

func getAccessToken(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	encoder := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	appID := params.Get("appId")

	if len(appID) <= 0 {
		encoder.Encode(map[string]interface{}{
			"error": "missing appId",
		})
		return
	}

	tokenHolder, err := pool.Get(appID)
	if err != nil {
		log.Printf("[%s]: %v", appID, err)
		encoder.Encode(map[string]interface{}{
			"error": "get token failed",
		})
		return
	}

	encoder.Encode(
		map[string]interface{}{
			"access_token": tokenHolder.AccessToken.Token,
			"expires_in":   tokenHolder.ExpiresIn().Seconds(),
		})
}

func newServer(addr string) *http.Server {
	m := &APIServerMux{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		mux:    http.NewServeMux(),
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/token", getAccessToken)

	m.mux.Handle("/", r)

	h := &http.Server{Addr: addr, Handler: m}

	return h

}

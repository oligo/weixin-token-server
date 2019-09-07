package main

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

// TokenStore defines the access token storage interface
type TokenStore interface {
	Load(appID string) (*AccessToken, error)
	Save(tokens ...AccessToken) error
}

//DiskStore stores token in file
type DiskStore struct {
	file  string
	mutex *sync.Mutex
}

// NewDiskStore creates a new disk store
func NewDiskStore(file string) TokenStore {
	return &DiskStore{
		file:  file,
		mutex: &sync.Mutex{},
	}
}

// Load loads token from file
func (d *DiskStore) Load(appID string) (*AccessToken, error) {
	tokenMap, err := d.loadAll()
	if err != nil {
		return nil, err
	}

	token, ok := tokenMap[appID]
	if !ok {
		return nil, errors.New("access token not found for " + appID)
	}

	return &token, nil
}

// Save persists token to file
func (d *DiskStore) Save(tokens ...AccessToken) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	tokenMap, err := d.loadAll()
	if err != nil {
		return err
	}

	for _, token := range tokens {
		tokenMap[token.AppID] = token
	}

	file, err := os.OpenFile(d.file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer file.Close()

	json.NewEncoder(file).Encode(tokenMap)

	return nil
}

func (d *DiskStore) loadAll() (map[string]AccessToken, error) {
	file, err := os.OpenFile(d.file, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	tokenMap := make(map[string]AccessToken)
	json.NewDecoder(file).Decode(&tokenMap)

	return tokenMap, nil
}

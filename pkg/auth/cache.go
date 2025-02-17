package auth

import (
	//cache "github.com/patrickmn/go-cache"
	"k8s.io/client-go/dynamic"
	"time"
)

type cacheManger struct {
	expiration time.Duration
	clientset  dynamic.Interface
	//authzCache *cache.Cache
}

func NewCache(expiration time.Duration) Cache {
	return &cacheManger{
		expiration: expiration,
		//authzCache: cache.New(expiration, 3*time.Minute),
	}
}

func (m *cacheManger) GetAuthorize(key string) (bool, error) {
	return false, nil
}

func (m *cacheManger) SetAuthorize(key string, val interface{}, d time.Duration) error {
	return nil
}

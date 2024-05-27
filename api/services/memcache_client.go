package services

import (
	"encoding/binary"
	"errors"
	"net/mail"
	"odo24_mobile_backend/config"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

var client *memcache.Client

func getMemcachedClient() *memcache.Client {
	cfg := config.GetInstance()
	if client == nil {
		client = memcache.New(cfg.Memcache.Addr)
	}
	return client
}

func newMemcachedClient() *memcache.Client {
	if client != nil {
		client.Close()
	}

	cfg := config.GetInstance()
	client = memcache.New(cfg.Memcache.Addr)
	return client
}

func AddEmailCodeConfirmation(email *mail.Address, code uint16) error {
	rawCode := make([]byte, 2)
	binary.LittleEndian.PutUint16(rawCode, code)

	cacheKey := strings.Replace(email.Address, "@", ".", -1)

	item := memcache.Item{
		Key:        cacheKey,
		Value:      rawCode,
		Expiration: 60 * 10,
	}

	memc := getMemcachedClient()
	for i := 0; i < 2; i++ {
		err := memc.Add(&item)
		if err != nil {
			if errors.Is(err, memcache.ErrNoServers) || errors.Is(err, memcache.ErrServerError) {
				newMemcachedClient()
				continue
			}
			return err
		}
		break
	}

	return nil
}

func GetEmailCodeConfirmation(email *mail.Address) (item *memcache.Item, err error) {
	memc := getMemcachedClient()
	cacheKey := strings.Replace(email.Address, "@", ".", -1)
	item, err = memc.Get(cacheKey)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func DeleteEmailCodeConfirmation(email *mail.Address) error {
	memc := getMemcachedClient()
	cacheKey := strings.Replace(email.Address, "@", ".", -1)
	return memc.Delete(cacheKey)
}

package services

import (
	"encoding/binary"
	"errors"
	"net/mail"
	"strconv"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/mil-ast/odo24_mobile_backend/config"
)

const confirmCodeExpiration = 60 * 10
const MaxConfirmCodeAttempts = 5
const loginAttemptExpiration = 60 * 15
const MaxLoginAttempts = 5

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

func getEmailCacheKey(email *mail.Address) string {
	return strings.Replace(email.Address, "@", ".", -1)
}

func getConfirmCodeAttemptsKey(email *mail.Address) string {
	return getEmailCacheKey(email) + ":attempts"
}

func AddEmailCodeConfirmation(email *mail.Address, code uint32) error {
	rawCode := make([]byte, 4)
	binary.LittleEndian.PutUint32(rawCode, code)

	item := memcache.Item{
		Key:        getEmailCacheKey(email),
		Value:      rawCode,
		Expiration: confirmCodeExpiration,
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

	return ResetConfirmCodeAttempts(email)
}

func GetEmailCodeConfirmation(email *mail.Address) (item *memcache.Item, err error) {
	memc := getMemcachedClient()
	item, err = memc.Get(getEmailCacheKey(email))
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
	if err := memc.Delete(getEmailCacheKey(email)); err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
		return err
	}
	return DeleteConfirmCodeAttempts(email)
}

func GetConfirmCodeAttempts(email *mail.Address) (int, error) {
	memc := getMemcachedClient()
	item, err := memc.Get(getConfirmCodeAttemptsKey(email))
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return 0, nil
		}
		return 0, err
	}
	n, err := strconv.Atoi(string(item.Value))
	if err != nil {
		return 0, nil
	}
	return n, nil
}

func IncrementConfirmCodeAttempts(email *mail.Address) (int, error) {
	memc := getMemcachedClient()

	attempts, err := GetConfirmCodeAttempts(email)
	if err != nil {
		return 0, err
	}
	attempts++

	err = memc.Set(&memcache.Item{
		Key:        getConfirmCodeAttemptsKey(email),
		Value:      []byte(strconv.Itoa(attempts)),
		Expiration: confirmCodeExpiration,
	})
	if err != nil {
		return 0, err
	}
	return attempts, nil
}

func ResetConfirmCodeAttempts(email *mail.Address) error {
	memc := getMemcachedClient()
	err := memc.Delete(getConfirmCodeAttemptsKey(email))
	if err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
		return err
	}
	return nil
}

func DeleteConfirmCodeAttempts(email *mail.Address) error {
	return ResetConfirmCodeAttempts(email)
}

func getLoginAttemptKey(login string) string {
	return "login_attempts:" + strings.ToLower(strings.Replace(login, "@", ".", -1))
}

func GetLoginAttempts(login string) (int, error) {
	memc := getMemcachedClient()
	item, err := memc.Get(getLoginAttemptKey(login))
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return 0, nil
		}
		return 0, err
	}
	n, err := strconv.Atoi(string(item.Value))
	if err != nil {
		return 0, nil
	}
	return n, nil
}

func IncrementLoginAttempts(login string) (int, error) {
	memc := getMemcachedClient()

	attempts, err := GetLoginAttempts(login)
	if err != nil {
		return 0, err
	}
	attempts++

	err = memc.Set(&memcache.Item{
		Key:        getLoginAttemptKey(login),
		Value:      []byte(strconv.Itoa(attempts)),
		Expiration: loginAttemptExpiration,
	})
	if err != nil {
		return 0, err
	}
	return attempts, nil
}

func ResetLoginAttempts(login string) error {
	memc := getMemcachedClient()
	err := memc.Delete(getLoginAttemptKey(login))
	if err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
		return err
	}
	return nil
}

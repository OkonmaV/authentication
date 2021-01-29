package main

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedStruct struct {
	Location string
}
type cacheItem struct {
	Key   string
	Value []byte
}

var memc = MemcachedStruct{Location: "127.0.0.1:11211"}

func main() {
	mc := memcache.New(memc.Location)

	key1, err := getMD5hash("login1")
	check(err)
	key11, err := getMD5hash("pass1")
	check(err)

	key2, err := getMD5hash("login2")
	check(err)
	key22, err := getMD5hash("pass2")
	check(err)

	user1 := &cacheItem{
		Key:   key1 + key11,
		Value: []byte("1 Name1 Surname1"),
	}

	user2 := &cacheItem{
		Key:   key2 + key22,
		Value: []byte("2 Name2 Surname2"),
	}

	err = mc.Set(&memcache.Item{Key: user1.Key, Value: user1.Value})
	check(err)

	err = mc.Set(&memcache.Item{Key: user2.Key, Value: user2.Value})
	check(err)
	return

}
func getMD5hash(str string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(str))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func check(err error) {
	if err != nil {
		println(err)
	}
}

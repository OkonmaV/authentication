package main

import (
	"github.com/bradfitz/gomemcache/memcache"
)

func (cache *MemcachedStruct) Valid(login, pass string) error {
	hashlogin, err := GetMD5hash(login)
	if err != nil {
		return err
	}
	//fmt.Println(hashlogin)
	hashpass, err := GetMD5hash(pass)
	if err != nil {
		return err
	}
	//fmt.Println(hashpass)
	mc := memcache.New(cache.Location)
	_, err = mc.Get(hashlogin + hashpass)
	return err
}

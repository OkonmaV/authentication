package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/dgrijalva/jwt-go"
)

type MemcachedStruct struct {
	Location string
}

type Claims struct {
	Uid     string
	Name    string
	Surname string
	IP      string
	jwt.StandardClaims
}

var memc = MemcachedStruct{Location: "127.0.0.1:11211"}
var jwtKey = []byte("secure_key")

func handler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return
	}
	login := r.Form["login"]
	pass := r.Form["pass"]

	err = memc.Valid(login[0], pass[0])
	if err != nil {
		fmt.Fprintf(w, "wrong login or pass")
		return
	}

	hashlogin, err := GetMD5hash(login[0])
	if err != nil {
		fmt.Println(err)
	}
	hashpass, err := GetMD5hash(pass[0])
	if err != nil {
		fmt.Println(err)
	}

	mc := memcache.New(memc.Location)
	cacheItem, err := mc.Get(hashlogin + hashpass)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = CreateCookie(w, r, cacheItem.Value)

	http.Redirect(w, r, r.Header.Get("Referer"), 302)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

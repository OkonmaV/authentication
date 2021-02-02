package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type foo struct {
	Login string `json:"login"`
}
type UserInfo struct {
	Name string `json:"name"`
}
type Claims struct {
	Login string
	Name  string
	//	IP    string
	jwt.StandardClaims
}

var jwtKey = "secure_key"

func handler(w http.ResponseWriter, r *http.Request) {

	bar := &foo{}

	if r.Body == nil {
		fmt.Println("Empty Body") //todo
		return
	}
	err := json.NewDecoder(r.Body).Decode(bar)
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	reqUserInfoBody, err := json.Marshal(bar)
	reqUserInfo, err := http.NewRequest("GET", "http://givememycookie", strings.NewReader(string(reqUserInfoBody))) //todo
	reqUserInfo.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	respUserInfo, err := client.Get("http://givemeinfo") //todo
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	defer func() {
		err := respUserInfo.Body.Close()

		if err != nil {
			fmt.Println(err) //todo
			return
		}
	}()

	if respUserInfo.StatusCode == http.StatusOK {

		userinfo := &UserInfo{}
		expTime := time.Now().Add(10 * time.Hour)

		err := json.NewDecoder(r.Body).Decode(userinfo)
		if err != nil {
			fmt.Println(err) //todo
			return
		}

		claims := &Claims{
			Login: bar.Login,
			Name:  userinfo.Name,
		}

		jwtTokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtKey)
		if err != nil {
			fmt.Println(err) //todo
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "koki",
			Value:   jwtTokenString,
			Expires: expTime,
		})
		return
	}
	//todo: bad status code in respUserInfo
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8084", nil))
}

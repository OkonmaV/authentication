package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	hashLogin := r.URL.Query().Get("l")
	claims := &Claims{Login: hashLogin}
	jwtTokenString, err := claims.GetJWT([]byte(jwtKey))
	if err != nil {
		fmt.Println(err) //todo
		return
	}

	reqUserInfo, err := http.NewRequest("GET", "http://givememyinfo?p=[name]", nil) //todo
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	reqUserInfo.AddCookie(&http.Cookie{
		Name:  "koki",
		Value: jwtTokenString,
	})

	client := &http.Client{}
	respUserInfo, err := client.Do(reqUserInfo)
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

		err := json.NewDecoder(respUserInfo.Body).Decode(userinfo)
		if err != nil {
			fmt.Println(err) //todo
			return
		}

		claims := &Claims{
			Login: hashLogin,
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

func (claims *Claims) GetJWT(key []byte) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}

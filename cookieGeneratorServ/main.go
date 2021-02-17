package main

import (
	"encoding/base64"
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
	Login  string
	Avatar string
	//	IP    string
	jwt.StandardClaims
}
type configs struct {
	jwtKey []byte
}

func (cfg *configs) handler(w http.ResponseWriter, r *http.Request) {

	hashLogin := r.URL.Query().Get("l")
	claims := &Claims{Login: hashLogin, Avatar: EncodeBase64(hashLogin)}
	jwtTokenString, err := claims.GetJWT(cfg.jwtKey)
	if err != nil {
		fmt.Println(err) //todo
		return
	}

	/*reqUserInfo, err := http.NewRequest("GET", "http://givememyinfo?p=[name]", nil) //todo
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
	userinfo := &UserInfo{}
	err := json.NewDecoder(respUserInfo.Body).Decode(userinfo)
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	*/
	expTime := time.Now().Add(10 * time.Hour)

	http.SetCookie(w, &http.Cookie{
		Name:    "koki",
		Value:   jwtTokenString,
		Expires: expTime,
	})
	w.WriteHeader(http.StatusOK)
}

func main() {
	cfg := &configs{jwtKey: []byte("secure_key")}
	http.HandleFunc("/", cfg.handler)
	log.Fatal(http.ListenAndServe(":8084", nil))
}

func (claims *Claims) GetJWT(key []byte) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}

func EncodeBase64(data string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(data))
}

func DecodeBase64(data string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(data)
}

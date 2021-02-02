package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
)

type Claims struct {
	Login string
	Name  string
	IP    string
	jwt.StandardClaims
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
var ctx = context.Background()

type configs struct {
	jwtKey      []byte
	redisClient *redis.Client
}

func (conf *configs) handler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return
	}
	// ------ Read & check recieved login ------
	login := r.Form["login"]
	if !IsEmailValid(login[0]) {
		fmt.Println(login[0] + "oopsie email") //todo
		return
	}

	hashLogin, err := GetMD5hash(login[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	// ------ Get password from redis ------
	hashRedisValue, err := conf.redisClient.Get(ctx, hashLogin).Result()
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	// ------ Read & check recieved password ------
	pass := r.Form["pass"]
	if len(pass[0]) < 8 {
		fmt.Println("short password") //todo
		return
	}

	hashPass, err := GetMD5hash(pass[0])
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	if hashPass != hashRedisValue {
		fmt.Println("wrong password") //todo
		return
	}
	// ------ Create request to userinfo ------
	reqUserInfo, err := http.NewRequest("POST", "http://127.0.0.1:8081/givememycookie", strings.NewReader("")) //todo
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	// ------ Create claims for jwt ------
	ip := r.Header.Get("X-Forwarded-For")
	claims := &Claims{
		Login: hashLogin,
		IP:    ip,
	}
	jwtReqTokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(conf.jwtKey)
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	// ------ Add cookie to request and execute request ------
	reqUserInfo.Header.Set("Cookie", "name=koki; count="+jwtReqTokenString)

	client := &http.Client{}
	respUserInfo, err := client.Do(reqUserInfo) // Close()?
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	// ------ Get cookie with userinfo and send it to user ------
	respCookies := respUserInfo.Cookies()
	if err != nil {
		fmt.Println(err) //todo
		return
	}

	http.SetCookie(w, respCookies[0])

	http.Redirect(w, r, r.Header.Get("Referer"), 302)
}

func main() {
	var jwtkey = []byte("secure_key")
	redisClient := redis.NewClient((&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}))
	rds := &configs{jwtKey: jwtkey, redisClient: redisClient}

	http.HandleFunc("/", rds.handler)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func IsEmailValid(email string) bool {
	if len(email) < 6 && len(email) > 30 {
		return false
	}
	if !emailRegex.MatchString(email) {
		return false
	}
	parts := strings.Split(email, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return false
	}
	return true
}

func GetMD5hash(str string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(str))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

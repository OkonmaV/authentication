package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
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
	//	IP    string
	jwt.StandardClaims
}

type UserInfo struct {
	Name string `json:"name"`
}
type foo struct {
	Login string `json:"login"`
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
	//------ Get cookie from reAuth ------
	reqReAuthBody, err := json.Marshal(&foo{
		Login: hashLogin,
	})
	reqReAuth, err := http.NewRequest("GET", "http://givememycookie", strings.NewReader(string(reqReAuthBody))) //todo
	reqReAuth.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	respReAuth, err := client.Do(reqReAuth) // Close()?
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	defer func() {
		err := respReAuth.Body.Close()

		if err != nil {
			fmt.Println(err) //todo
			return
		}
	}()
	//------ Get and set cookie ------
	respCookies := respReAuth.Cookies()
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

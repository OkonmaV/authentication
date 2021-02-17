package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tarantool/go-tarantool"
)

type Claims struct {
	Login string
	Name  string
	//Surname string
	//Avatar string
	//	IP    string
	jwt.StandardClaims
}

type UserInfo struct {
	Name string `json:"name"`
}
type foo struct {
	Login string `json:"login"`
}
type Tuple struct {
	Login string
	Pass  string
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
var ctx = context.Background()

type configs struct {
	jwtKey        []byte
	tarantoolConn *tarantool.Connection
}

func (cfg *configs) handler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		fmt.Println(err) //todo
		return
	}

	// ------ Read & check recieved login ------

	login := r.FormValue("login")
	if !IsEmailValid(login) {
		fmt.Println(login + "oopsie email") //todo
		return
	}

	hashLogin, err := GetMD5(login)
	if err != nil {
		fmt.Println(err)
		return
	}

	// ------ Get password from tarantool ------
	var tarantoolResTuples []Tuple
	err = cfg.tarantoolConn.SelectTyped("main", "primary", 0, 1, tarantool.IterEq, []interface{}{hashLogin}, &tarantoolResTuples)
	if err != nil || len(tarantoolResTuples) == 0 {
		fmt.Println(err) //todo wrong login
		return
	}

	// ------ Read & check recieved password ------

	pass := r.FormValue("pass")
	if len(pass) < 8 {
		fmt.Println("short password") //todo
		return
	}

	hashPass, err := GetMD5(pass)
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	if hashPass != tarantoolResTuples[0].Pass {
		fmt.Println("wrong password") //todo
		return
	}

	//------ Get cookie from cookieGenerator ------

	client := &http.Client{}
	respCookieGen, err := client.Get("http://givememycookie?l=" + hashLogin) // Close()?
	if err != nil {
		fmt.Println(err) //todo
		return
	}

	if respCookieGen.StatusCode != http.StatusOK {
		fmt.Println("bad rest from cookieGen") //todo
		return
	}

	//------ Get and set cookie ------

	respCookies := respCookieGen.Cookies()
	if err != nil {
		fmt.Println(err) //todo
		return
	}

	http.SetCookie(w, respCookies[0])

	http.Redirect(w, r, r.Header.Get("Referer"), 302) //todo
}

func main() {
	var jwtkey = []byte("secure_key")
	conn, err := tarantool.Connect("127.0.0.1:3301", tarantool.Opts{
		User:          "admin",
		Pass:          "password",
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 4,
	})
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println(err) //todo
			return
		}
	}()
	cfg := *&configs{jwtKey: jwtkey, tarantoolConn: conn}

	http.HandleFunc("/", cfg.handler)
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

func GetMD5(str string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(str))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
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

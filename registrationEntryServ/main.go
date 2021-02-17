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
	"time"

	"github.com/beevik/guid"
	"github.com/tarantool/go-tarantool"
)

type configs struct {
	tarantoolConn *tarantool.Connection
}
type Tuple struct {
	Login string
	Pass  string
}

var ctx = context.Background()
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (cfg *configs) handler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		fmt.Println(err) //todo
		return
	}

	mail := r.FormValue("mail")
	if !IsEmailValid(mail) {
		fmt.Println("mail broken") //todo
		return
	}

	mailHash, err := GetMD5(mail)
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	var tarantoolResTuples []Tuple
	err = cfg.tarantoolConn.SelectTyped("main", "primary", 0, 1, tarantool.IterEq, []interface{}{mailHash}, &tarantoolResTuples)
	if err != nil || len(tarantoolResTuples) != 0 {
		fmt.Println(err) //todo
		return
	}

	uid := guid.New()
	_, err = cfg.tarantoolConn.Insert("limbo", []interface{}{mail, uid.String()})
	if err != nil {
		fmt.Println(err) //todo
		return
	}

	//SEND MAIL WITH GUID
}

func main() {
	connTrntl, err := tarantool.Connect("localhost:3301", tarantool.Opts{
		// User:          "admin",
		// Pass:          "password",
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 4,
	})
	if err != nil {
		fmt.Println("tarantool", err) //todo
		return
	}
	defer func() {
		if connTrntl.Close() != nil {
			fmt.Println(err) //todo
			return
		}
	}()

	cfg := *&configs{tarantoolConn: connTrntl}

	http.HandleFunc("/", cfg.handler)
	log.Fatal(http.ListenAndServe(":8086", nil))
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

package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/beevik/guid"
	"github.com/tarantool/go-tarantool"
)

type configs struct {
	tarantoolConn *tarantool.Connection
}
type Tuple struct {
	Login string
	Guid  string
}

func (cfg *configs) handler(w http.ResponseWriter, r *http.Request) {

	uid := r.URL.Query().Get("g")
	mail := r.URL.Query().Get("m")

	if !guid.IsGuid(uid) {
		fmt.Println("broken guid") //todo
		return
	}

	var tarantoolResTuples []Tuple
	err := cfg.tarantoolConn.SelectTyped("limbo", "primary", 0, 1, tarantool.IterEq, []interface{}{mail}, &tarantoolResTuples)
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	if uid != tarantoolResTuples[0].Guid {
		fmt.Println("wrong guid") //todo
		return
	}
	r.Header.Add("X-Foo", EncodeBase64(mail))
	http.ServeFile(w, r, "form_registration.html")
}

func main() {
	connTrntl, err := tarantool.Connect("127.0.0.1:3301", tarantool.Opts{
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
		if connTrntl.Close() != nil {
			fmt.Println(err) //todo
			return
		}
	}()

	cfg := *&configs{tarantoolConn: connTrntl}

	http.HandleFunc("/", cfg.handler)
	log.Fatal(http.ListenAndServe(":8087", nil))
}

func EncodeBase64(data string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(data))
}

func DecodeBase64(data string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(data)
}

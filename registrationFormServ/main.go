package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
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
	Guid  string
}

const form = `
        <form action="/registry" method="POST">
			<input type="hidden" name="guid" value="%%">
            <input placeholder="name" name="name">
            <input placeholder="surname" name="surname">
            <input placeholder="password" type="password" name="password">
            <input type="submit" value="registry">
        </form>
`

func (cfg *configs) handler(w http.ResponseWriter, r *http.Request) {

	uid := r.URL.Query().Get("g")

	if !guid.IsGuid(uid) {
		fmt.Println("broken guid") //todo
		return
	}

	var tarantoolResTuples []Tuple
	err := cfg.tarantoolConn.SelectTyped("limbo", "secondary", 0, 1, tarantool.IterEq, []interface{}{uid}, &tarantoolResTuples)
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	if len(tarantoolResTuples) == 0 {
		fmt.Println("not found") //todo
		return
	}
	if uid != tarantoolResTuples[0].Guid {
		fmt.Println("wrong guid") //todo
		return
	}
	fmt.Println(EncodeBase64(tarantoolResTuples[0].Login))
	w.Write([]byte(strings.ReplaceAll(form, "%%", tarantoolResTuples[0].Guid)))
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

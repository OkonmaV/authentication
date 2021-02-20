package main

import (
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

const form = `<form action="http://localhost:8088" method="POST">
	<input type="hidden" name="guid" value="%some%">
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var tarantoolResTuples []Tuple
	err := cfg.tarantoolConn.SelectTyped("limbo", "secondary", 0, 1, tarantool.IterEq, []interface{}{uid}, &tarantoolResTuples)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(err) //todo
		return
	}
	if len(tarantoolResTuples) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("not found") //todo
		return
	}
	_, err = w.Write([]byte(strings.ReplaceAll(form, "%some%", tarantoolResTuples[0].Guid)))
	if err != nil {

	}
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

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tarantool/go-tarantool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type configs struct {
	tarantoolConn *tarantool.Connection
	mongoConn     *mongo.Client
	mongoColl     *mongo.Collection
}

var ctx = context.Background()

type Tuple struct {
	Login string
	Check bool
}

type UserInfo struct {
	Login   string `bson:"_id"`
	Name    string `bson:"name"`
	Surname string `bson:"surname"`
}

func (cfg *configs) handler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		fmt.Println(err) //todo
		return
	}

	queryValues := r.URL.Query()
	userinfo := &UserInfo{Login: queryValues.Get("l"), Name: queryValues.Get("n"), Surname: queryValues.Get("s")}

	var tarantoolResTuples []Tuple
	err = cfg.tarantoolConn.SelectTyped("limbo", "primary", 0, 1, tarantool.IterEq, []interface{}{userinfo.Login}, &tarantoolResTuples)
	if err != nil || len(tarantoolResTuples) == 0 {
		fmt.Println(err) //todo
		return
	}

	if tarantoolResTuples[0].Check {
		opts := options.Update().SetUpsert(true)
		filter := bson.D{{"_id", userinfo.Login}}
		update := bson.D{{"$set", bson.D{{"name", userinfo.Name}, {"surname", userinfo.Surname}}}}

		_, err := cfg.mongoColl.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			fmt.Println(err) //todo
			return
		}
		_, err = cfg.tarantoolConn.Upsert("main", []interface{}{queryValues.Get("l")}, []interface{}{[]interface{}{queryValues.Get("l"), queryValues.Get("p")}}) //Insert("main", []interface{}{queryValues.Get("l"), queryValues.Get("p")})
		if err != nil {
			fmt.Println(err) //todo
			return
		}
		_, err = cfg.tarantoolConn.Delete("limbo", "primary", []interface{}{queryValues.Get("l")})
	}
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

	connMng, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Println(err) //todo
		return
	}
	defer func() {
		if connMng.Disconnect(ctx) != nil {
			fmt.Println(err)
		}
	}()

	collectionMng := connMng.Database("main").Collection("users")

	cfg := *&configs{tarantoolConn: connTrntl, mongoConn: connMng, mongoColl: collectionMng}

	http.HandleFunc("/", cfg.handler)
	log.Fatal(http.ListenAndServe(":8084", nil))
}

package main

import (
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {

}
func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8084", nil))
}

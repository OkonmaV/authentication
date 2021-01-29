package main

import (
	"log"
	"net/http"
	"time"
)

// Quit serv
func handler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "koki",
		MaxAge:  -1,
		Expires: time.Now().Add(-5 * time.Hour),
	})
	http.Redirect(w, r, "/", 302)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8083", nil))
}

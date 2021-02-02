package main

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func CreateCookie(w http.ResponseWriter, r *http.Request, byteinfo []byte) error {
	ip := r.Header.Get("X-Real-IP")
	expTime := time.Now().Add(10 * time.Hour)
	info := strings.Fields(string(byteinfo))
	claims := &Claims{
		Uid:     info[0],
		Name:    info[1],
		Surname: info[2],
		IP:      ip,
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtKey)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "koki",
		Value:   tokenString,
		Expires: expTime,
	})
	return nil
}

func GetMD5hash(str string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(str))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

func SetAuthCookie(w http.ResponseWriter, tok *oauth2.Token) {
	jsonTok, err := json.Marshal(tok)
	HandleError(w, err)

	b64Tok := string(ToB64([]byte(jsonTok)))

	cookie := &http.Cookie{
		Name:  "Auth",
		Value: b64Tok,
		// Secure: true,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
}

func GetAuthToken(cookie *http.Cookie, w http.ResponseWriter) *oauth2.Token {
	var tok *oauth2.Token
	jsonTok := Decode64([]byte(cookie.Value))
	for jsonTok[len(jsonTok)-1] == 0 {
		jsonTok = jsonTok[:len(jsonTok)-1]
	}
	err := json.Unmarshal(jsonTok, &tok)
	HandleError(w, err)
	return tok
}

func ToB64(str []byte) []byte {
	encLen := base64.StdEncoding.EncodedLen(len(str))
	b64 := make([]byte, encLen)
	base64.StdEncoding.Encode(b64, str)
	return b64
}

func Decode64(x []byte) []byte {
	decLen := base64.StdEncoding.DecodedLen(len(x))
	str := make([]byte, decLen)
	base64.StdEncoding.Decode(str, x)
	return str
}

func HandleError(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalln(err)
	}
}
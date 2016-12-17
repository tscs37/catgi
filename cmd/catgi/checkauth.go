package main

import (
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

type handlerCheckToken struct {
	next http.Handler
}

func newHandlerCheckToken(nextHandler http.Handler) http.Handler {
	return &handlerCheckToken{
		next: nextHandler,
	}
}

func (h *handlerCheckToken) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If no users are configured, disable authentication
	if !(curCfg.Users == nil || len(curCfg.Users) == 0) {
		cookie, err := r.Cookie("auth")
		if err == http.ErrNoCookie {
			w.WriteHeader(401)
			fmt.Fprint(w, "No Token")
			return
		}
		token := cookie.Value
		t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			if token.Method.Alg() != jwt.SigningMethodHS512.Alg() {
				return nil, jwt.ErrInvalidKeyType
			}
			return []byte(curCfg.HMACKey), nil
		})
		if err != nil {
			w.WriteHeader(401)
			fmt.Fprint(w, "401 - Not Authorized")
			return
		}
		if !t.Valid {
			w.WriteHeader(401)
			fmt.Fprint(w, "401 - Not Authorized")
			return
		}
	}

	h.next.ServeHTTP(w, r)
}

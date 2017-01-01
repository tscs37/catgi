package main

import (
	"fmt"
	"net/http"

	"context"

	"git.timschuster.info/rls.moe/catgi/logger"
	jwt "github.com/dgrijalva/jwt-go"
)

type handlerCheckToken struct {
	next http.Handler
	lazy bool
}

// lazy: If set to true it will check the token and return a Header
// with the username only but not abort with a 401 on failure
func newHandlerCheckToken(lazy bool, nextHandler http.Handler) http.Handler {
	return &handlerCheckToken{
		next: nextHandler,
		lazy: lazy,
	}
}

func (h *handlerCheckToken) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("httpCheckAuth:"+fmt.Sprint(h.lazy), r.Context())
	var decodedClaims jwt.MapClaims
	var decodedToken *jwt.Token

	// If no users are configured, disable authentication
	if curCfg.Users == nil || len(curCfg.Users) == 0 {
		log.Warn("No users configured, skipping auth-check")
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user", "anonymous")

		r = r.WithContext(ctx)
		h.next.ServeHTTP(w, r)
		return
	}
	
	{
		cookie, err := r.Cookie("auth")
		if err == http.ErrNoCookie {
			h.abortLogin(w, r)
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
			log.Warn("Error on JWT Decode: ", err)
			h.abortLogin(w, r)
			return
		}
		if !t.Valid {
			log.Warn("JWT expired or invalid")
			h.abortLogin(w, r)
			return
		}
		if claims, ok := t.Claims.(jwt.MapClaims); !ok {
			h.abortLogin(w, r)
			log.Warn("JWT not in standard format")
			return
		} else {
			log.Debug("Saving Claims for Lazy Auth")
			decodedClaims = claims
			decodedToken = t
		}
	}

	if decodedClaims != nil {
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user", (decodedClaims)["sub"])

		r = r.WithContext(ctx)

		w.Header().Add("X-Catgi-Logged-In-As", (decodedClaims)["sub"].(string))
	} else {
		log.Error("JWT had nil or invalid claims")
		log.Debug("JWT dump: ", decodedToken, decodedClaims)
	}

	h.next.ServeHTTP(w, r)
}

func (h *handlerCheckToken) abortLogin(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("httpCheckAuth:abort", r.Context())
	if !h.lazy {
		w.WriteHeader(401)
		fmt.Fprint(w, "401 - Not Authorized")
	} else {
		log.Info("Request is using Lazy-Auth, not aborting Auth")
		h.next.ServeHTTP(w, r)
	}
}

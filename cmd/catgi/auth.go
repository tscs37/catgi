package main

import (
	"fmt"
	"net/http"
	"time"

	"git.timschuster.info/rls.moe/catgi/logger"
	"git.timschuster.info/rls.moe/catgi/snowflakes"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/hlandau/passlib"
)

type handlerServeAuth struct{}

func newHandlerServeAuth() http.Handler {
	return &handlerServeAuth{}
}

func (h *handlerServeAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("postAuth", r.Context())

	log.Info("Incoming Auth Request")

	log.Debug("Parsing Form")
	err := r.ParseForm()
	if err != nil {
		log.Warn("Could not parse form: ", err)
		w.WriteHeader(500)
		fmt.Fprint(w, "Error while parsing incoming data")
		return
	}

	user := r.FormValue("user")
	pass := r.FormValue("pass")

	log.Debug("Checking password")

	for _, v := range curCfg.Users {
		if v.Username == user {
			log.Debug("Found user, verifying")
			_, err = passlib.Verify(pass, v.PassHash)
			if err != nil {
				log.Warn("Wrong password attempt for user ", user)
				w.WriteHeader(401)
				fmt.Fprint(w, "401 - Not Authorized")
				return
			} else {
				claimflake, err := snowflakes.NewSnowflake()
				if err != nil {
					log.Warn("Could not generate claim flake: ", err)
					w.WriteHeader(401)
					fmt.Fprint(w, "401 - Not Authorized")
					return
				}
				claims := &jwt.StandardClaims{
					ExpiresAt: time.Now().AddDate(0, 2, 0).Unix(),
					Issuer:    "catgi.rls.moe",
					IssuedAt:  time.Now().Unix(),
					NotBefore: time.Now().Unix(),
					Id:        claimflake,
					Subject:   user,
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

				tokenString, err := token.SignedString([]byte(curCfg.HMACKey))

				if err != nil {
					log.Error("Error on auth: ", err)
					w.WriteHeader(401)
					fmt.Fprint(w, "401 - Not Authorized")
					return
				}

				log.Debug("Setting auth cooie: ", tokenString)

				http.SetCookie(w, &http.Cookie{
					Name:     "auth",
					Value:    tokenString,
					Expires:  time.Now().AddDate(0, 1, 0),
					Secure:   true,
					HttpOnly: true,
				})

				fmt.Fprintf(w, "Logged in as %s.\nReturn to main page to upload files now.", user)
				return
			}
		}
	}
	log.Debug("Could not find user ", user)
	w.WriteHeader(401)
	fmt.Fprintf(w, "401 - Not Authorized")
}

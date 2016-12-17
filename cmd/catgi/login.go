package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"git.timschuster.info/rls.moe/catgi/logger"
)

type handlerServeLogin struct{}

func newHandlerServeLogin() http.Handler {
	return &handlerServeLogin{}
}

func (h *handlerServeLogin) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("serveLogin", r.Context())
	dat, err := ioutil.ReadFile("./login.html")
	if err != nil {
		log.Error("Could not load file from disk: ", err)
		rw.WriteHeader(404)
		fmt.Fprint(rw, "login.html not found")
		return
	}
	rw.WriteHeader(200)
	rw.Header().Add("Content-Type", "application/html")
	rw.Write(dat)
}

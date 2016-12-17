package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"git.timschuster.info/rls.moe/catgi/logger"
)

type handlerServeSite struct{}

func newHandlerServeSite() http.Handler {
	return &handlerServeSite{}
}
func (h *handlerServeSite) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("serverIndex", r.Context())
	dat, err := ioutil.ReadFile("./index.html")
	if err != nil {
		log.Error("Could not load file from disk: ", err)
		rw.WriteHeader(404)
		fmt.Fprint(rw, "index.html not found")
		return
	}
	rw.WriteHeader(200)
	rw.Header().Add("Content-Type", "application/html")
	rw.Write(dat)
}

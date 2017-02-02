package main

import (
	"fmt"
	"net/http"

	rice "github.com/GeertJohan/go.rice"

	"git.timschuster.info/rls.moe/catgi/logger"
)

type handlerServeResources struct {
	rice *rice.Box
}

func newHandlerServeResources() http.Handler {
	return &handlerServeResources{
		rice: (&rice.Config{
			LocateOrder: []rice.LocateMethod{
				rice.LocateWorkingDirectory,
				rice.LocateFS,
				rice.LocateEmbedded,
			},
		}).MustFindBox("./resources"),
	}
}
func (h *handlerServeResources) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("serverIndex", r.Context())
	log.Info("Loading file from disk: ", r.RequestURI)
	dat, err := h.rice.Bytes(r.URL.String())
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

package main

import (
	"fmt"
	"net/http"

	rice "github.com/GeertJohan/go.rice"

	"git.timschuster.info/rls.moe/catgi/logger"
)

type handlerServeSite struct {
	rice rice.Config
}

func newHandlerServeSite() http.Handler {
	return &handlerServeSite{
		rice: rice.Config{
			LocateOrder: []rice.LocateMethod{
				rice.LocateWorkingDirectory,
				rice.LocateFS,
				rice.LocateEmbedded,
			},
		},
	}
}
func (h *handlerServeSite) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("serverIndex", r.Context())
	dat, err := h.rice.MustFindBox("./resources").Bytes("index.html")
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

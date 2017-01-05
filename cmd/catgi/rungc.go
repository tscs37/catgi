package main

import (
	"encoding/json"
	"net/http"

	"git.timschuster.info/rls.moe/catgi/logger"
)

type handlerRunGC struct{}

func newHandlerRunGC() http.Handler {
	return &handlerRunGC{}
}

func (h *handlerRunGC) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("runGC", r.Context())

	log.Info("Starting GC")

	files, err := curBe.RunGC(r.Context())

	if err != nil {
		w.WriteHeader(500)
		dat, err := json.Marshal(err)
		if err != nil {
			log.Error("Error on error encode: ", err)
			w.Write([]byte("Critical Server Error"))
			return
		}
		w.Write(dat)
		return
	}

	dat, err := json.Marshal(files)

	if err != nil {
		w.WriteHeader(500)
		err_dat, err := json.Marshal(err)
		if err != nil {
			log.Error("Error on error encode: ", err)
			w.Write([]byte("Critical Server Error"))
			return
		}
		w.Write(err_dat)
		return
	}

	w.Write(dat)
}

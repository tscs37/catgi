package main

import (
	"encoding/json"
	"net/http"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
)

type handlerRunGC struct {
	backend common.Backend
}

func newHandlerRunGC(b common.Backend) http.Handler {
	return &handlerRunGC{
		backend: b,
	}
}

func (h *handlerRunGC) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("runGC", r.Context())

	log.Info("Starting GC")

	files, err := h.backend.RunGC(r.Context())

	if err != nil {
		w.WriteHeader(500)
		var dat []byte
		dat, err = json.Marshal(err)
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
		errDat, err := json.Marshal(err)
		if err != nil {
			log.Error("Error on error encode: ", err)
			w.Write([]byte("Critical Server Error"))
			return
		}
		w.Write(errDat)
		return
	}

	w.Write(dat)
}

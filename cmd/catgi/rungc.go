package main

import (
	"encoding/json"
	"net/http"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
	"git.timschuster.info/rls.moe/catgi/utils"
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

	r = r.WithContext(utils.PutHTTPIntoContext(r, r.Context()))

	// <- BEGIN BACKEND INTERACTION ->
	files, err := h.backend.RunGC(r.Context())
	// -> END BACKEND INTERACTION

	if err != nil && !common.IsHTTPOption(err) {
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
	} else if common.IsHTTPOption(err) {
		httpopt := err.(common.ErrorHTTPOptions)
		httpopt.PassOverHTTP(w)
		if httpopt.WantsTakeover() {
			httpopt.HTTPTakeover(r, w, r.Context())
			return
		}
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

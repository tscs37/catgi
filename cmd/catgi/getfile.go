package main

import (
	"fmt"
	"net/http"

	"bitbucket.org/taruti/mimemagic"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/gorilla/mux"
)

type handlerServeGet struct{}

func newHandlerServeGet() http.Handler {
	return &handlerServeGet{}
}

func (h *handlerServeGet) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("getFile", r.Context())

	log.Debug("Loading Flake")
	vars := mux.Vars(r)
	flake := vars["flake"]
	if len(flake) == 0 {
		log.Warn("Form contained no flake")
		rw.WriteHeader(500)
		fmt.Fprint(rw, "Missing flake")
		return
	}

	log.Debug("Loading File from Backend")
	f, err := curBe.Get(flake, r.Context())
	if err != nil {
		log.Warn("File error on backend: ", err)
		rw.WriteHeader(404)
		fmt.Fprint(rw, "Could not find file")
		return
	}

	log.Debug("Writing out response")

	mimetype := ""

	if f.ContentType == "" {
		if len(f.Data) > 1024 {
			mimetype = mimemagic.Match("image/png", f.Data[:1024])
		} else {
			mimetype = mimemagic.Match("image/png", f.Data)
		}
	} else {
		mimetype = f.ContentType
	}

	rw.Header().Add("Content-Disposition", "inline; filename="+flake+f.FileExtension)
	rw.Header().Add("Content-Type", mimetype)
	_, err = rw.Write(f.Data)
	if err != nil {
		log.Errorf("Error on store: %s", err)
	}

	return
}

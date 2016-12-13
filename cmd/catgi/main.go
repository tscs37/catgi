package main

import (
	"fmt"

	"net/http"

	"git.timschuster.info/rls.moe/catgi/backend"
	_ "git.timschuster.info/rls.moe/catgi/backend/b2"
	_ "git.timschuster.info/rls.moe/catgi/backend/buntdb"
	"git.timschuster.info/rls.moe/catgi/config"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/InVisionApp/rye"
	"github.com/gorilla/mux"
)

var (
	curBe backend.Backend
)

func main() {
	ctx := logger.NewLoggingContext()
	logger.SetLoggingLevel("info", ctx)

	log := logger.LogFromCtx("main", ctx)

	conf, err := config.LoadConfig("./conf.json")

	log.Info("Starting Backend")
	be, err := backend.NewBackend(conf.Backend.Name, conf.Backend.Params, ctx)
	if err != nil {
		log.Errorf("Error: %s", err)
		return
	}
	curBe = be
	log.Infof("Loaded '%s' Backend Driver", be.Name())
	mwHandler := rye.NewMWHandler(rye.Config{})

	router := mux.NewRouter()
	router.Handle("/", mwHandler.Handle([]rye.Handler{
		serveGet,
	})).Methods("GET", "POST")

	http.ListenAndServe("[::1]:8080", router)
}

func injectLogToRequest(_ http.Response, r *http.Request) *rye.Response {
	return &rye.Response{
		Context: logger.InjectLogToContext(r.Context()),
	}
}

func serveGet(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log := logger.LogFromCtx("httpRequest", r.Context())

	fmt.Fprint(rw, "Hi")
	log.Info("HI!")
	curBe.Get("test", r.Context())
	return nil
}

package main

import (
	"fmt"
	"net/http"

	"git.timschuster.info/rls.moe/catgi/backend"
	_ "git.timschuster.info/rls.moe/catgi/backend/b2"
	_ "git.timschuster.info/rls.moe/catgi/backend/buntdb"
	_ "git.timschuster.info/rls.moe/catgi/backend/fcache"
	"git.timschuster.info/rls.moe/catgi/config"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/gorilla/mux"
)

var (
	curBe  backend.Backend
	curCfg config.Configuration
)

func main() {
	ctx := logger.NewLoggingContext()
	ctx = logger.SetLoggingLevel("DEBUG", ctx)

	log := logger.LogFromCtx("main", ctx)

	var err error
	curCfg, err = config.LoadConfig("./conf.json")

	log.Info("Starting Backend")
	be, err := backend.NewBackend(curCfg.Backend.Name, curCfg.Backend.Params, ctx)
	if err != nil {
		log.Errorf("Error: %s", err)
		return
	}
	curBe = be
	log.Infof("Loaded '%s' Backend Driver", be.Name())
	router := mux.NewRouter()
	router.Handle("/file/{flake}",
		newHandlerInjectLog(
			newHandlerServeGet(),
		),
	).Methods("GET")

	router.Handle("/file",
		newHandlerInjectLog(
			newHandlerCheckToken(
				newHandlerServePost(),
			),
		),
	).Methods("POST")

	router.Handle("/",
		newHandlerInjectLog(
			newHandlerServeSite(),
		),
	).Methods("GET")

	router.Handle("/login",
		newHandlerInjectLog(
			newHandlerServeLogin(),
		),
	).Methods("GET")

	router.Handle("/auth",
		newHandlerInjectLog(
			newHandlerServeAuth(),
		),
	).Methods("POST")

	listenOn := curCfg.HTTPConf.ListenOn +
		fmt.Sprintf(":%d", curCfg.HTTPConf.Port)
	log.Info("Starting HTTP Service on ", listenOn)

	err = http.ListenAndServe(
		listenOn,
		router,
	)
	if err != nil {
		log.Fatal(err)
	}
}

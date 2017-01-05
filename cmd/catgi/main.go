package main

import (
	"fmt"
	"net/http"
	"os"

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
	var err error
	if len(os.Args) < 2 {
		print("Using default config\n")
		curCfg.Backend = config.DriverConfig{
			Name: "buntdb",
			Params: map[string]interface{}{
				"file": ":memory:",
			},
		}
		curCfg.HMACKey = ""
		curCfg.HTTPConf = config.HTTPConfig{
			ListenOn: "[::1]",
			Port:     8080,
		}
		curCfg.LogLevel = "debug"
		curCfg.Users = []config.UserConfig{}

	} else {
		fmt.Printf("%s\n", os.Args)
		curCfg, err = config.LoadConfig(os.Args[1])
		if err != nil {
			fmt.Printf("Config not valid: %s\n", err)
			return
		}
	}

	ctx := logger.NewLoggingContext()
	ctx = logger.SetLoggingLevel(curCfg.LogLevel, ctx)
	log := logger.LogFromCtx("main", ctx)

	log.Info("Starting Backend")
	be, err := backend.NewBackend(curCfg.Backend.Name, curCfg.Backend.Params, ctx)
	if err != nil {
		log.Errorf("Error: %s", err)
		return
	}
	curBe = be
	log.Infof("Loaded '%s' Backend Driver", be.Name())

	piwik := newHandlerPiwik(curCfg.Piwik.Base, curCfg.Piwik.ID,
		curCfg.Piwik.Enable, curCfg.Piwik.IgnoreErrors)

	router := mux.NewRouter()
	{
		fileGetHandler := newHandlerInjectLog(
			piwik(
				newHandlerCheckToken(true,
					newHandlerServeGet(),
				),
			),
		)

		router.StrictSlash(false).Handle("/file/{flake}",
			fileGetHandler,
		).Methods("GET")

		router.StrictSlash(false).Handle("/f/{flake}",
			fileGetHandler,
		).Methods("GET")

		router.StrictSlash(false).Handle("/f/{flake}/",
			fileGetHandler,
		).Methods("GET")

		router.Handle("/f/{flake}/{name}.{ext}",
			fileGetHandler,
		).Methods("GET")
	}

	router.Handle("/file",
		newHandlerInjectLog(
			piwik(
				newHandlerCheckToken(false,
					newHandlerServePost(),
				),
			),
		),
	).Methods("POST")

	router.Handle("/gc",
		newHandlerInjectLog(
			newHandlerCheckToken(false,
				newHandlerRunGC(),
			),
		),
	).Methods("GET")

	router.Handle("/",
		newHandlerInjectLog(
			piwik(
				newHandlerCheckToken(true,
					newHandlerServeSite(),
				),
			),
		),
	).Methods("GET")

	router.Handle("/login",
		newHandlerInjectLog(
			piwik(
				newHandlerServeLogin(),
			),
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

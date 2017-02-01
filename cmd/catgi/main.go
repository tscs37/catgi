package main

import (
	"fmt"
	"net/http"
	"os"

	"os/signal"
	"syscall"

	"net"

	"git.timschuster.info/rls.moe/catgi/backend"
	_ "git.timschuster.info/rls.moe/catgi/backend/b2"
	_ "git.timschuster.info/rls.moe/catgi/backend/buntdb"
	_ "git.timschuster.info/rls.moe/catgi/backend/fcache"
	_ "git.timschuster.info/rls.moe/catgi/backend/localfs"
	_ "git.timschuster.info/rls.moe/catgi/backend/s3"
	"git.timschuster.info/rls.moe/catgi/config"
	"git.timschuster.info/rls.moe/catgi/logger"
	"git.timschuster.info/rls.moe/catgi/utils"
	"github.com/gorilla/mux"
)

var (
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
	log.Infof("Loaded '%s' Backend Driver", be.Name())

	piwik := newHandlerPiwik(curCfg.Piwik.Base, curCfg.Piwik.ID,
		curCfg.Piwik.Enable, curCfg.Piwik.IgnoreErrors)

	router := mux.NewRouter()
	{
		fileGetHandler := newHandlerInjectLog(
			piwik(
				newHandlerCheckToken(true,
					newHandlerServeGet(be),
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
					newHandlerServePost(be),
				),
			),
		),
	).Methods("POST")

	router.Handle("/gc",
		newHandlerInjectLog(
			newHandlerCheckToken(false,
				newHandlerRunGC(be),
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

	router.PathPrefix("/res/").Handler(http.StripPrefix("/res/",
		newHandlerInjectLog(
			newHandlerCheckToken(true,
				newHandlerServeResources(),
			),
		),
	)).Methods("GET")

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
	)

	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		log.Fatal(err)
		return
	}

	stpLst := utils.Handle(listener)

	closeServ := func(stpLst *utils.StoppableListener) {
		log.Print("Terminating Webserver")
		stpLst.Stop <- true
		log.Print("Waiting for connections to finish")
		utils.WaitFor(func() bool { return stpLst.ConnCount.Get() == 0 })
		log.Print("Server Stopped and All Connection Closed")
	}

	go func(sig chan os.Signal, stpLst *utils.StoppableListener) {
		for {
			v := <-sig
			log.Print("Received System Signal")
			switch v {
			case os.Interrupt:
				closeServ(stpLst)
				os.Exit(1)
			case syscall.SIGHUP:
				closeServ(stpLst)
				// 0-Code Exit on HUP
				os.Exit(0)
			case syscall.SIGINT:
				closeServ(stpLst)
				os.Exit(1)
			case syscall.SIGTERM:
				closeServ(stpLst)
				os.Exit(1)
			case syscall.SIGKILL:
				closeServ(stpLst)
				os.Exit(1)
			default:
			}
		}
	}(sigChan, stpLst)

	err = http.Serve(
		stpLst,
		router,
	)
	if err != nil {
		log.Fatal(err)
	}
}

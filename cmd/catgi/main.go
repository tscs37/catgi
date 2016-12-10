package main

import (
	"time"

	"git.timschuster.info/rls.moe/catgi/backend"
	_ "git.timschuster.info/rls.moe/catgi/backend/b2"
	_ "git.timschuster.info/rls.moe/catgi/backend/buntdb"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/config"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/Sirupsen/logrus"
)

func main() {
	ctx := logger.NewLoggingContext()
	logger.SetLoggingLevel("debug", ctx)

	log := logger.LogFromCtx("main", ctx)

	conf, err := config.LoadConfig("./conf.json")

	log.Info("Starting Backend")
	be, err := backend.NewBackend(conf.Backend.Name, conf.Backend.Params, ctx)
	if err != nil {
		log.Errorf("Error: %s", err)
		return
	}
	log.Infof("Loaded '%s' Backend Driver", be.Name())
	log.Info("Starting Index")
	idx, err := backend.NewIndex(conf.Index.Name, conf.Index.Params, ctx)
	if err != nil {
		log.Errorf("Error: %s", err)
		return
	}
	log.Infof("Loaded '%s' Index Driver", idx.Name())

	log.Info("Loading index")
	err = be.LoadIndex(idx, ctx)
	if err != nil {
		log.Errorf("Error: %s", err)
	}

	log.Info("Storing Hello World")
	_, f, err := idx.Put(types.File{
		Data:     []byte("Hello World"),
		Public:   true,
		DeleteAt: time.Unix(0, 0),
	}, be, ctx)
	if err != nil {
		log.Errorf("Error: %s", err)
		return
	}
	log.Infof("Stored at /%s/", f.Flake)

	log.Infof("Checking if file exists")

	err = be.Exists(f.Flake, ctx)
	if err != nil {
		log.Infof("Result: %s", err)
	}
	err = be.Exists("doesnotexist", ctx)
	if err != nil {
		log.Infof("Result: %s", err)
	}

	_, f, err = idx.Get(types.File{
		Flake: f.Flake,
	}, be, ctx)
	if err != nil {
		logrus.Errorf("Error: %s", err)
		return
	}
	log.WithField("file", "/"+f.Flake+"/").Infof("File contains %s'", f.Data)

	err = be.StoreIndex(idx, ctx)
	if err != nil {
		log.Error("Error on store: %s", err)
	}
}

package main

import (
	"time"

	"net/http"

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
	logger.SetLoggingLevel("info", ctx)

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

	log.Info("Storing Hello World")
	_, f, err := idx.Put(types.File{
		Data:     []byte("Hello World"),
		Public:   true,
		DeleteAt: time.Now().UTC().Add(1 * time.Hour),
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

	log.Infof("Checking if doesnotexist exits :3")
	err = be.Exists("doesnotexist", ctx)
	if err != nil {
		log.Infof("Result: %s", err)
	}

	log.Infof("Retrieving file")
	_, f, err = idx.Get(types.File{
		Flake: f.Flake,
	}, be, ctx)
	if err != nil {
		logrus.Errorf("Error: %s", err)
		return
	}
	log.WithField("file", "/"+f.Flake+"/").Infof("File contains '%s'", f.Data)

	log.Infof("Cleaning up bucket")
	err = be.CleanUp(ctx)
	if err != nil {
		logrus.Errorf("Error: %s", err)
		return
	}
}

func serveGet(w http.ResponseWriter, r *http.Request) {

}

package main

import (
	"time"

	"git.timschuster.info/rls.moe/catgi/backend"
	_ "git.timschuster.info/rls.moe/catgi/backend/b2"
	_ "git.timschuster.info/rls.moe/catgi/backend/buntdb"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/config"
	"git.timschuster.info/rls.moe/catgi/snowflakes"
	"github.com/Sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	conf, err := config.LoadConfig("./conf.json")
	flake, err := snowflakes.NewSnowflake()
	if err != nil {
		logrus.Errorf("Error: %s", err)
		return
	}
	logrus.Info("Out: ", flake)
	logrus.Infof("Drivers: %s", backend.InstalledDrivers())
	logrus.Info("Starting Backend")
	be, err := backend.NewBackend(conf.Backend.Name, conf.Backend.Params)
	if err != nil {
		logrus.Errorf("Error: %s", err)
		return
	}
	logrus.Infof("Loaded '%s' Backend Driver", be.Name())
	logrus.Info("Starting Index")
	idx, err := backend.NewIndex(conf.Index.Name, conf.Index.Params)
	if err != nil {
		logrus.Errorf("Error: %s", err)
		return
	}
	logrus.Infof("Loaded '%s' Index Driver", idx.Name())

	logrus.Info("Loading index")
	err = be.LoadIndex(idx)
	if err != nil {
		logrus.Errorf("Error: %s", err)
	}

	logrus.Info("Store Index")
	err = be.StoreIndex(idx)
	if err != nil {
		logrus.Errorf("Error: %s", err)
	}

	logrus.Info("Loading Index again")
	err = be.LoadIndex(idx)
	if err != nil {
		logrus.Errorf("Error: %s", err)
	}

	logrus.Info("Storing Hello World")
	_, f, err := idx.Put(types.File{
		Data:     []byte("Hello World"),
		Public:   true,
		DeleteAt: time.Unix(0, 0),
	}, be)
	if err != nil {
		logrus.Errorf("Error: %s", err)
		return
	}
	logrus.Infof("Stored at /%s/", f.Flake)

	_, f, err = idx.Get(types.GetRequest{
		Flake: f.Flake,
	}, be)
	if err != nil {
		logrus.Errorf("Error: %s", err)
		return
	}
	logrus.Infof("File at /%s/ contains: '%s'", f.Flake, f.Data)
	logrus.Debugf("Raw Data: %#v", f)
}

package main

import (
	"git.timschuster.info/rls.moe/catgi/backend"
	_ "git.timschuster.info/rls.moe/catgi/backend/b2"
	_ "git.timschuster.info/rls.moe/catgi/backend/buntdb"
	"git.timschuster.info/rls.moe/catgi/config"
	"git.timschuster.info/rls.moe/catgi/snowflakes"
	"github.com/Sirupsen/logrus"
)

func main() {
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
}

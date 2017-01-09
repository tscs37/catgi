package buntdb

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/buntdb"
)

type buntConfig struct {
	File string `mapstructure:"file"`
	// NoAutoTTL disables automatic TTL and requires manual GC
	// which may reduce CPU usage.
	NoAutoTTL bool `mapstructure:"no_auto_expire"`
}

func init() {
	backend.NewDriver("buntdb", NewBuntDBBackend)
}

func NewBuntDBBackend(params map[string]interface{}, ctx context.Context) (types.Backend, error) {
	log := logger.LogFromCtx(bePackagename+".New", ctx)
	var config = &buntConfig{
		File:      ":memory:",
		NoAutoTTL: false,
	}
	{
		log.Debug("Loading Config")
		decConf := &mapstructure.DecoderConfig{
			ErrorUnused:      true,
			WeaklyTypedInput: true,
			ZeroFields:       false,
			Result:           config,
		}
		decoder, err := mapstructure.NewDecoder(decConf)
		if err != nil {
			return nil, err
		}

		err = decoder.Decode(params)
		if err != nil {
			return nil, err
		}
		log.Debug("Config Loading Complete")
	}
	log.Debug("Opening DB")
	db, err := buntdb.Open(config.File)
	if err != nil {
		log.Error("Error on DB open, returning: ", err)
		return nil, err
	}
	log.Debug("Driver initialized.")
	return &BuntDBBackend{
		db:      db,
		file:    config.File,
		autoTTL: !config.NoAutoTTL,
	}, nil
}

package buntdb

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/buntdb"
)

type buntConfig struct {
	File string `mapstructure:"file"`
}

func init() {
	backend.NewDriver("buntdb", NewBuntDBBackend)
}

func NewBuntDBBackend(params map[string]interface{}, ctx context.Context) (types.Backend, error) {
	var config = &buntConfig{}
	{
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
	}
	db, err := buntdb.Open(config.File)
	if err != nil {
		return nil, err
	}
	return &BuntDBBackend{db: db}, nil
}
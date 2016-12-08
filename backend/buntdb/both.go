package buntdb

import (
	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/buntdb"
)

type buntConfig struct {
	File string `mapstructure:"file"`
}

const (
	indexFileDeleteAt  = "index-file-deleteat"
	indexFilePublic    = "index-file-public"
	indexFileCreatedAt = "index-file-createdat"
)

func init() {
	backend.NewDriver("buntdb", NewBuntDBIndex)
	backend.NewDriver("buntdb", NewBuntDBBackend)
}

func NewBuntDBIndex(params map[string]interface{}) (types.Index, error) {
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

	err = db.CreateIndex(indexFileDeleteAt, "/meta/file/*/delete_at", buntdb.IndexInt)
	if err != nil {
		return nil, err
	}
	err = db.CreateIndex(indexFileCreatedAt, "/meta/file/*/created_at", buntdb.IndexInt)
	if err != nil {
		return nil, err
	}
	err = db.CreateIndex(indexFilePublic, "/meta/file/*/public", buntdb.IndexString)
	if err != nil {
		return nil, err
	}
	return &BuntDBIndex{db: db}, nil
}

func NewBuntDBBackend(params map[string]interface{}) (types.Backend, error) {
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

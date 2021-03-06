package localfs

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/mitchellh/mapstructure"
)

const (
	packageName = "localfs"
)

func init() {
	backend.NewDriver("localfs", NewLocalFSBackend)
}

func NewLocalFSBackend(params map[string]interface{}, ctx context.Context) (common.Backend, error) {
	log := logger.LogFromCtx(packageName+".New", ctx)

	var config = &LocalFSBackend{
		Root:         "/localfs/",
		AbsoluteRoot: false,
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

	return config, config.pingFS()
}

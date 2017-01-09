package localfs

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/logger"
)

const (
	packageName = "localfs"
)

type localfsConfig struct {
	// The Root Path of the localfs backend. This will be
	// forced relative unless AbsoluteRoot is set.
	Root string `mapstructure:"root"`
	// If set true, the root will be trated as absolute path
	// instead of forcing a relative path.
	AbsoluteRoot bool `mapstructure:"abs_root"`
}

func init() {
	backend.NewDriver("localfs", NewLocalFSBackend)
}

func NewLocalFSBackend(params map[string]interface{}, ctx context.Context) (types.Backend, error) {
	log := logger.LogFromCtx(packageName+".New", ctx)
}

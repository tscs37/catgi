package fcache

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/bluele/gcache"
	"github.com/mitchellh/mapstructure"
)

type FCacheConfig struct {
	// Underlying Backend Driver
	Driver string `mapstructure:"driver"`
	// Underlying Backend Driver Configuration
	DriverConfig map[string]interface{} `mapstructure:"params"`
	// Number of Entries in the Cache
	Size int `mapstructure:"cache_size"`
	// If set to true, then any upload will hit the cache
	// and be uploaded in the background. May cause issues on
	// heavy traffic and errors cannot be propagated to the user
	// Files may disappear without warning.
	AsyncUpload bool `mapstructure:"async_upload"`
}

const driverName = "fcache"
const packageName = "backend/fcache"

func init() {
	backend.NewDriver(driverName, NewFCacheBackend)
}

// FCache is a WIP Cache Structure
type FCache struct {
	underlyingBackend common.Backend
	cache             gcache.Cache
	asyncUpload       bool
}

func NewFCacheBackend(params map[string]interface{}, ctx context.Context) (common.Backend, error) {
	var config = &FCacheConfig{}
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
	ub, err := backend.NewBackend(config.Driver, config.DriverConfig, ctx)
	if err != nil {
		return nil, err
	}

	intCache := gcache.New(config.Size).ARC().Build()
	return &FCache{
		underlyingBackend: ub,
		cache:             intCache,
		asyncUpload:       config.AsyncUpload,
	}, nil
}

func (n *FCache) Name() string { return driverName }

func (n *FCache) Upload(flake string, file *common.File, ctx context.Context) error {
	if file == nil {
		return common.ErrorSerializationFailure
	}

	if file.Flake != flake {
		file.Flake = flake
	}
	if !n.asyncUpload {
		err := n.underlyingBackend.Upload(flake, file, ctx)
		if err != nil {
			return err
		}
		n.cache.Set(flake, file)
		return nil
	} else if n.asyncUpload {
		n.cache.Set(flake, file)
		go func() {
			err := n.underlyingBackend.Upload(flake, file, ctx)
			if err != nil {
				// If the upload fails, evict the file from cache
				n.cache.Remove(flake)
			}
		}()
		return nil
	}
	return common.ErrorNotImplemented
}

func (n *FCache) Exists(flake string, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Get", ctx)
	if _, err := n.cache.Get(flake); err == nil {
		log.Debug("File in cache, exists if not stale")
		return nil
	}
	log.Debug("File not in cache, asking backend")
	return n.underlyingBackend.Exists(flake, ctx)
}

func (n *FCache) Get(flake string, ctx context.Context) (*common.File, error) {
	log := logger.LogFromCtx(packageName+".Get", ctx)
	log.Debug("Checking if file is in cache")
	if val, err := n.cache.Get(flake); err == nil {
		log.Debug("Checking if cache contains file (it should)")
		if f, ok := val.(*common.File); ok {
			log.Debug("Answering request from cache")
			if f.Data == nil {
				f.Data = []byte{}
			}
			return f, nil
		}
		log.Error("Cache did not contain file")
	}
	log.Info("Cache Miss, loading from backend")
	f, err := n.underlyingBackend.Get(flake, ctx)
	if err != nil {
		return nil, err
	}
	n.cache.Set(flake, f)
	return f, nil
}

func (n *FCache) Delete(flake string, ctx context.Context) error {
	defer func(flake string, ctx context.Context) {
		i, err := n.cache.Get(flake)
		if err != nil {
			return
		}
		n.cache.Remove(i)
	}(flake, ctx)
	return n.underlyingBackend.Delete(flake, ctx)
}

func (n *FCache) ListGlob(ctx context.Context, prefix string) ([]*common.File, error) {
	return n.underlyingBackend.ListGlob(ctx, prefix)
}

// FCache will clear any GC'd files from the cache.
// This makes unpleasant things die.
func (n *FCache) RunGC(ctx context.Context) ([]common.File, error) {
	files, err := n.underlyingBackend.RunGC(ctx)
	// After a GC we purge the cache since the actual
	// contents in the backend may have changed drastically.
	n.cache.Purge()
	return files, err
}

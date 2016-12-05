package backend

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"github.com/kurin/blazer/b2"
	"github.com/mitchellh/mapstructure"
)

const driverName = "b2"

type B2Backend struct {
	client *b2.Client
	config *b2Config
}

type b2Config struct {
	AccountID     string `mapstructure:"acc-id"`
	AccountSecret string `mapstructure:"acc-sec"`
	IndexBucket   string `mapstructure:"idx-bucket"`
	DataBucket    string `mapstructure:"dat-bucket"`
}

func init() {
	if err := backend.NewDriver(driverName, NewB2Backend); err != nil {
		panic(err)
	}
}

func NewB2Backend(params map[string]interface{}) (types.Backend, error) {
	var config = &b2Config{}
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

	var client *b2.Client
	{
		var err error
		client, err = b2.NewClient(context.Background(), config.AccountID, config.AccountSecret)
		if err != nil {
			return nil, err
		}
	}

	return &B2Backend{
		client: client,
		config: config,
	}, nil
}

func (b2 *B2Backend) Name() string { return driverName }

func (b2 *B2Backend) Upload(flake string, file *types.File) error { return nil }

func (b2 *B2Backend) Exists(flake string) error { return nil }

func (b2 *B2Backend) Get(flake string) (*types.File, error) {
	return &types.File{}, nil
}

func (b2 *B2Backend) Delete(flake string) error { return nil }

func (b2 *B2Backend) LoadIndex(i types.Index) error { return nil }

func (b2 *B2Backend) StoreIndex(i types.Index) error { return nil }

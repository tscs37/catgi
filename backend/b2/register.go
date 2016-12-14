package backend

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"github.com/kurin/blazer/b2"
	"github.com/mitchellh/mapstructure"
)

// B2Backend represents a initialized B2 Storage Backend connection
type B2Backend struct {
	client     *b2.Client
	config     *b2Config
	dataBucket *b2.Bucket
}

type b2Config struct {
	AccountID     string `mapstructure:"acc-id"`
	AccountSecret string `mapstructure:"acc-sec"`
	DataBucket    string `mapstructure:"dat-bucket"`
}

func init() {
	if err := backend.NewDriver(driverName, NewB2Backend); err != nil {
		panic(err)
	}
}

// NewB2Backend parses the incoming config from the mapstring
// and preloads the account buckets
func NewB2Backend(params map[string]interface{}, ctx context.Context) (types.Backend, error) {
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

	var datBuck *b2.Bucket
	{
		bucket, err := client.Bucket(ctx, config.DataBucket)
		if err != nil {
			return nil, err
		}
		datBuck = bucket
	}

	return &B2Backend{
		client:     client,
		config:     config,
		dataBucket: datBuck,
	}, nil
}

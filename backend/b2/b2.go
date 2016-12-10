package backend

import (
	"context"
	"io"

	"bytes"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/Sirupsen/logrus"
	"github.com/kurin/blazer/b2"
	blazer "github.com/kurin/blazer/b2"
	"github.com/mitchellh/mapstructure"
)

const driverName = "b2"
const packageName = "backend/b2"

// B2Backend represents a initialized B2 Storage Backend connection
type B2Backend struct {
	client     *blazer.Client
	config     *b2Config
	dataBucket *blazer.Bucket
	indxBucket *blazer.Bucket
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

	var client *blazer.Client
	{
		var err error
		client, err = blazer.NewClient(context.Background(), config.AccountID, config.AccountSecret)
		if err != nil {
			return nil, err
		}
	}

	var datBuck, idxBuck *blazer.Bucket
	{
		bucket, err := client.Bucket(ctx, config.DataBucket)
		if err != nil {
			return nil, err
		}
		datBuck = bucket
		bucket, err = client.Bucket(ctx, config.IndexBucket)
		if err != nil {
			return nil, err
		}
		idxBuck = bucket
	}

	return &B2Backend{
		client:     client,
		config:     config,
		dataBucket: datBuck,
		indxBucket: idxBuck,
	}, nil
}

// Name returns the current drive Name
func (b *B2Backend) Name() string { return driverName }

// Upload writes to the object in B2
func (b *B2Backend) Upload(flake string, file *types.File, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Upload", ctx)
	obj := b.dataBucket.Object("file/" + flake)
	log.Debug("Opening new Writer")
	w := obj.NewWriter(ctx)
	defer w.Close()
	log.Debug("Creating new File Buffer")
	buf := bytes.NewBuffer(file.Data)
	log.Debug("Copying Data to B2")
	if _, err := io.Copy(w, buf); err != nil {
		log.Errorf("Error while uploading: %s", err)
		return err
	}
	return nil
}

// Exists checks if the object exists in B2
func (b *B2Backend) Exists(flake string, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Exists", ctx)
	log.Debug("Getting context and object")
	obj := b.dataBucket.Object("file/" + flake)
	log.Debug("Got object, reading attrs")
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		logrus.Error("Attr not read, returning")
		return err
	}
	if attrs.Status != b2.Uploaded {
		return types.ErrorFileNotExist
	}
	log.Debug("Object is ok, returning")
	return nil
}

// Get reads the B2 File from the backend
func (b *B2Backend) Get(flake string, ctx context.Context) (*types.File, error) {
	obj := b.dataBucket.Object("file/" + flake)
	{
		attr, err := obj.Attrs(ctx)
		if err != nil {
			return nil, err
		}
		if attr.Status != blazer.Uploaded || attr.Size == 0 {
			return nil, types.ErrorFileNotExist
		}
	}
	r := obj.NewReader(ctx)
	defer r.Close()
	r.ConcurrentDownloads = 2
	buf := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(buf, r); err != nil {
		return nil, err
	}
	return &types.File{
		Data: buf.Bytes(),
	}, nil
}

// Delete empties the file on the B2 backend
func (b *B2Backend) Delete(flake string, ctx context.Context) error {
	obj := b.dataBucket.Object("file/" + flake)
	return obj.Delete(ctx)
}

// LoadIndex loads an index from idxBucket/index.blob
func (b *B2Backend) LoadIndex(i types.Index, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".LoadIndex", ctx)
	log.Debug("Loading Context and Object")
	obj := b.indxBucket.Object("idx.blob")
	log.Debug("Opening Reader")
	r := obj.NewReader(ctx)
	attr, err := obj.Attrs(ctx)
	if err != nil {
		return err
	}
	if attr.Status != blazer.Uploaded || attr.Size == 0 {
		log.Warn("Index has 0 size or is not uploaded yet, skipping...")
		return nil
	}
	log.Debugf("Loading file with %d bytes", attr.Size)
	defer r.Close()
	log.Debug("Buffer is nil")
	buf := bytes.NewBuffer([]byte{})
	log.Debug("Copying data...")
	if _, err := io.Copy(buf, r); err != nil {
		log.Error("Error when Loading: ", err)
		return err
	}
	log.Debug("Unserializing...")
	return i.Unserialize(buf.Bytes(), ctx)
}

// StoreIndex stores the index in idxBucket/index.blob
func (b *B2Backend) StoreIndex(i types.Index, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".StoreIndex", ctx)
	log.Debug("Loading Context and Object")
	obj := b.indxBucket.Object("idx.blob")
	log.Debug("Loading Writer")
	w := obj.NewWriter(ctx)
	defer w.Close()
	log.Debug("Serializing Index")
	dat, err := i.Serialize(ctx)
	if err != nil {
		return err
	}
	log.Debugf("Index Data is %d bytes", len(dat))
	buf := bytes.NewBuffer(dat)
	log.Debug("Writing")
	if _, err := io.Copy(w, buf); err != nil {
		log.Error("Error on Store: ", err)
		return err
	}
	log.Debug("Returning")
	return nil
}

// ListGlob returns a list of all files in the bucket
func (b *B2Backend) ListGlob(
	glob string, ictx context.Context) (
	[]*types.File, context.Context, error) {
	return nil, nil, nil
}

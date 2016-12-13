package backend

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"bytes"

	"encoding/json"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/kurin/blazer/b2"
	"github.com/mitchellh/mapstructure"
)

const driverName = "b2"
const packageName = "backend/b2"

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

// Name returns the current drive Name
func (b *B2Backend) Name() string { return driverName }

func dataName(flake string) string { return "file/" + splitName(flake) + "/data.bin" }

func metaName(flake string) string { return "file/" + splitName(flake) + "/meta.json" }

func (b *B2Backend) writeFile(name string, data []byte, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".writeFile", ctx).
		WithField("object", name).WithField("obj_len", len(data))
	obj := b.dataBucket.Object(name)
	log.Debug("Opening new Writer")
	w := obj.NewWriter(ctx)
	defer w.Close()
	log.Debug("Creating new Data Buffer")
	buf := bytes.NewBuffer(data)
	if n, err := io.Copy(w, buf); err != nil {
		log.Error("Error while uploading: ", err)
		return err
	} else {
		log.Debugf("Wrote %d bytes", n)
		return nil
	}
}

func (b *B2Backend) deleteFile(name string, ctx context.Context) error {
	return b.dataBucket.Object(name).Delete(ctx)
}

func (b *B2Backend) pingFile(name string, ctx context.Context) (bool, *b2.Attrs, error) {
	log := logger.LogFromCtx(packageName+".pingFile", ctx).
		WithField("object", name)
	obj := b.dataBucket.Object(name)
	log.Debug("Loading object attributes")
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return false, nil, types.NewErrorFileNotExists(name, err)
	}

	if attrs.Status == b2.Uploaded {
		return true, attrs, nil
	}
	return false, attrs, nil
}

func (b *B2Backend) readFile(name string, ctx context.Context) ([]byte, error) {
	log := logger.LogFromCtx(packageName+".readFile", ctx).
		WithField("object", name)
	obj := b.dataBucket.Object(name)
	log.Debug("Loading object attributes")

	r := obj.NewReader(ctx)

	buffer := bytes.NewBuffer([]byte{})
	if n, err := io.Copy(buffer, r); err != nil {
		log.Error("Error while reading: ", err)
		return nil, err
	} else {
		log.Debugf("Read %d bytes", n)
		return buffer.Bytes(), nil
	}
}

func splitName(flakeStr string) string {
	flake := []rune(flakeStr)
	var out = []string{}
	skipSize := 2
	for true {
		if len(flake) > skipSize+1 {
			out = append(out, string(flake[0:skipSize]))
			flake = flake[skipSize:]
		} else {
			out = append(out, string(flake))
			return strings.Join(out, "/")
		}
	}
	panic("Should not terminate here")
}

// Upload writes to the object in B2
func (b *B2Backend) Upload(flake string, file *types.File, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Upload", ctx)
	log.Debug("Creating object '", flake, "'")
	log.Debug("Writing File Data")
	dataName := dataName(flake)
	metaName := metaName(flake)
	log.Debug("Writing to ", dataName)
	if err := b.writeFile(dataName, file.Data, ctx); err != nil {
		log.Error("Error writing data ", err)
		return err
	}
	log.Debug("Marshalling for ", metaName)
	metaFile := *file
	metaFile.Data = []byte{}
	dat, err := json.Marshal(metaFile)
	if err != nil {
		return err
	}
	log.Debug("Writing to ", metaName)
	if err := b.writeFile(metaName, dat, ctx); err != nil {
		log.Error("Error writing data ", err)
		return err
	}
	return nil
}

// Exists checks if the object exists in B2
func (b *B2Backend) Exists(flake string, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Exists", ctx)
	log.Debug("Getting context and object")
	dataName := dataName(flake)
	metaName := metaName(flake)
	exists, _, err := b.pingFile(dataName, ctx)
	if err != nil {
		return types.NewErrorFileNotExists(flake, err)
	}
	if !exists {
		return types.NewErrorFileNotExists(flake, nil)
	}
	exists, _, err = b.pingFile(metaName, ctx)
	if err != nil {
		return types.NewErrorFileNotExists(flake, err)
	}
	if !exists {
		return types.NewErrorFileNotExists(flake, nil)
	}
	return nil
}

// Get reads the B2 File from the backend
func (b *B2Backend) Get(flake string, ctx context.Context) (*types.File, error) {
	log := logger.LogFromCtx(packageName+".Exists", ctx).WithField("object", flake)
	var file = &types.File{}
	dataName := dataName(flake)
	metaName := metaName(flake)

	{
		log.Debug("Loading Meta File")
		dat, err := b.readFile(metaName, ctx)
		if err != nil {
			return nil, err
		}
		log.Debug("Unmarshalling Meta File")
		err = json.Unmarshal(dat, file)
		if err != nil {
			return nil, err
		}
		log.Debug("Checking Expiry Data: ", file.DeleteAt.Sub(time.Now().UTC()))
		if time.Now().UTC().After(file.DeleteAt) {
			log.Debug("Expired, deleting")
			err = b.deleteFile(metaName, ctx)
			if err != nil {
				log.Error("Error deleting metadata: ", err)
				return nil, err
			}
			err = b.deleteFile(dataName, ctx)
			if err != nil {
				log.Error("Error deleting data: ", err)
				return nil, err
			}
			return nil, errors.New("File expired")
		}
	}

	{
		dat, err := b.readFile(dataName, ctx)
		if err != nil {
			return nil, err
		}
		file.Data = dat
	}

	return file, nil
}

// Delete empties the file on the B2 backend
func (b *B2Backend) Delete(flake string, ctx context.Context) error {
	dataName := dataName(flake)
	metaName := metaName(flake)

	err := b.deleteFile(dataName, ctx)
	if err != nil {
		return err
	}

	err = b.deleteFile(metaName, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (b *B2Backend) CleanUp(ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".CleanUp", ctx)
	log.Debug("Scanning files...")
	var cur *b2.Cursor
	var compSearch = map[string]string{}
	for {
		objs, c, err := b.dataBucket.ListCurrentObjects(ctx, 1000, cur)
		if err != nil && err != io.EOF {
			return err
		}
		for _, obj := range objs {
			attr, err := obj.Attrs(ctx)
			if err != nil {
				log.Error("Error on ", obj, ": ", err)
			}
			log.Debug("Found object '", attr.Name, "'")
			isBin := strings.HasSuffix(attr.Name, "/data.bin")
			isMeta := strings.HasSuffix(attr.Name, "/meta.json")
			if !isBin && !isMeta {
				log.Warn("Found non-file object in bucket, deleting!")
				if err := obj.Delete(ctx); err != nil {
					log.Error("Non-Critical Error: ", err)
				}
			} else if isBin {
				stripped := strings.TrimSuffix(attr.Name, "/data.bin")
				if val, ok := compSearch[stripped+"/meta.json"]; ok {
					if val == "data.bin" {
						log.Debug("Found '"+stripped+"' companion: ", val)
						delete(compSearch, stripped+"/meta.json")
					} else {
						log.Error("Wrong companion: ", val, " for ", attr.Name, "/", stripped)
					}
				} else {
					compSearch[stripped+"/data.bin"] = "meta.json"
				}
			} else if isMeta {
				stripped := strings.TrimSuffix(attr.Name, "/meta.json")
				if val, ok := compSearch[stripped+"/data.bin"]; ok {
					if val == "meta.json" {
						log.Debug("Found '"+stripped+"' companion: ", val)
						delete(compSearch, stripped+"/data.bin")
					} else {
						log.Error("Wrong companion: ", val, " for ", attr.Name, "/", stripped)
					}
				} else {
					compSearch[stripped+"/meta.json"] = "data.bin"
				}
			}
		}
		if err == io.EOF {
			break
		}
		cur = c
	}
	for k, v := range compSearch {
		log.Warnf("Lone file: %s is missing %s", k, v)
		err := b.dataBucket.Object(k).Delete(ctx)
		if err != nil {
			log.Error("Non-Critical Error while deleting loner: ", err)
		} else {
			log.Info("Deleted Longer ", k)
		}
	}
	return nil
}

// ListGlob returns a list of all files in the bucket
func (b *B2Backend) ListGlob(
	glob string, ictx context.Context) (
	[]*types.File, context.Context, error) {
	return nil, nil, nil
}

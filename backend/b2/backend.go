package backend

import (
	"context"
	"io"
	"time"

	"encoding/json"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/kurin/blazer/b2"
)

const driverName = "b2"
const packageName = "backend/b2"
const skipSize = 2
const metaFormat = "json"

// Name returns the current drive Name
func (b *B2Backend) Name() string { return driverName }

// Upload writes to the object in B2
func (b *B2Backend) Upload(flake string, file *common.File, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Upload", ctx)
	log.Debug("Creating object '", flake, "'")
	file.CreatedAt = common.FromTime(time.Now().UTC())
	log.Debug("Writing File Data")
	dataName := common.DataName(flake, skipSize)
	metaName := common.MetaName(flake, skipSize, metaFormat)
	log.Debug("Writing to ", dataName)
	if err := b.writeFile(dataName, file.Data, ctx); err != nil {
		log.Error("Error writing data ", err)
		return err
	}
	log.Debug("Marshalling for ", metaName)
	oldData := file.Data
	file.Data = []byte{}
	dat, err := json.Marshal(file)
	if err != nil {
		return err
	}
	log.Debug("Writing to ", metaName)
	if err := b.writeFile(metaName, dat, ctx); err != nil {
		log.Error("Error writing data ", err)
		return err
	}
	file.Data = oldData
	return nil
}

// Exists checks if the object exists in B2
func (b *B2Backend) Exists(flake string, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Exists", ctx)
	log.Debug("Getting context and object")
	dataName := common.DataName(flake, skipSize)
	metaName := common.MetaName(flake, skipSize, metaFormat)
	exists, _, err := b.pingFile(dataName, ctx)
	if err != nil {
		return common.NewErrorFileNotExists(flake, err)
	}
	if !exists {
		return common.NewErrorFileNotExists(flake, nil)
	}
	exists, _, err = b.pingFile(metaName, ctx)
	if err != nil {
		return common.NewErrorFileNotExists(flake, err)
	}
	if !exists {
		return common.NewErrorFileNotExists(flake, nil)
	}
	return nil
}

// Get reads the B2 File from the backend
func (b *B2Backend) Get(flake string, ctx context.Context) (*common.File, error) {
	log := logger.LogFromCtx(packageName+".Exists", ctx).WithField("object", flake)
	var file = &common.File{}
	dataName := common.DataName(flake, skipSize)
	metaName := common.MetaName(flake, skipSize, metaFormat)

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
		if time.Now().UTC().After(file.DeleteAt.Time) {
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
			return nil, common.ErrorExpired
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
	dataName := common.DataName(flake, skipSize)
	metaName := common.MetaName(flake, skipSize, metaFormat)

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

// ListGlob returns a list of all files in the bucket
func (b *B2Backend) ListGlob(ctx context.Context, glob string) ([]*common.File, error) {
	log := logger.LogFromCtx(packageName+".ListGlob", ctx)
	files := []*common.File{}
	var cur *b2.Cursor
	for {
		objs, c, err := b.dataBucket.ListCurrentObjects(ctx, 1000, cur)
		if err != nil && err != io.EOF {
			return nil, err
		}
		for _, obj := range objs {
			if !common.IsMetaFile(obj.Name(), metaFormat) {
				continue
			}
			var dat []byte
			dat, err = b.readFile(obj.Name(), ctx)
			if err != nil {
				log.Error("Read error on glob: ", err)
			} else {
				var curFile = &common.File{}
				err = json.Unmarshal(dat, curFile)
				if err != nil {
					log.Error("Meta Unmarshal Error: ", err)
				} else {
					files = append(files, curFile)
				}
			}
		}
		if err == io.EOF {
			return files, nil
		}
		cur = c
	}
}

func (b *B2Backend) RunGC(ctx context.Context) ([]common.File, error) {
	return common.GenericGC(b, nil, nil, ctx)
}

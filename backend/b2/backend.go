package backend

import (
	"context"
	"time"

	"encoding/json"

	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/kurin/blazer/b2"
)

const driverName = "b2"
const packageName = "backend/b2"

// Name returns the current drive Name
func (b *B2Backend) Name() string { return driverName }

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
			return nil, types.ErrorExpired
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

// ListGlob returns a list of all files in the bucket
func (b *B2Backend) ListGlob(
	glob string, ictx context.Context) (
	[]*types.File, context.Context, error) {
	return nil, nil, nil
}

func (b *B2Backend) Publish(flake []string, name string, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Publish", ctx)
	if len(flake) > 10 {
		log.Error("Too many flakes for publishing")
		return types.ErrorQuotaExceeded
	}
	dat, err := json.Marshal(flake)
	if err != nil {
		log.Error("Could not marshall flake array")
		return err
	}

	err = b.writeFile(clpubName(name), dat, ctx)

	if err != nil {
		log.Error("Could not write to publisher index :", err)
		return err
	}
	return nil
}

func (b *B2Backend) Unpublish(name string, ctx context.Context) error {
	return b.deleteFile(clpubName(name), ctx)
}

func (b *B2Backend) Resolve(name string, ctx context.Context) ([]string, error) {
	log := logger.LogFromCtx(packageName+".Resolve", ctx)

	exists, attr, err := b.pingFile(name, ctx)
	if err != nil {
		log.Error("Could not read file from backend: ", err)
		return nil, err
	}
	if !(exists && attr.Status == b2.Uploaded) {
		log.Error("Could not read file from backend: exist check failed")
		return nil, types.NewErrorFileNotExists(name, nil)
	}

	dat, err := b.readFile(clpubName(name), ctx)

	if err != nil {
		log.Error("Could not read file from backend: ", err)
		return nil, err
	}

	var out = []string{}

	err = json.Unmarshal(dat, &out)
	return out, err
}

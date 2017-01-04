package backend

import (
	"context"
	"io"
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
	file.CreatedAt = types.FromTime(time.Now().UTC())
	log.Debug("Writing File Data")
	dataName := dataName(flake)
	metaName := metaName(flake)
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
func (b *B2Backend) ListGlob(ctx context.Context, glob string) ([]*types.File, error) {
	log := logger.LogFromCtx(packageName+".ListGlob", ctx)
	files := []*types.File{}
	var cur *b2.Cursor
	for {
		objs, c, err := b.dataBucket.ListCurrentObjects(ctx, 1000, cur)
		if err != nil && err != io.EOF {
			return nil, err
		}
		for _, obj := range objs {
			if !isMetaFile(obj.Name()) {
				continue
			}
			var dat []byte
			dat, err = b.readFile(obj.Name(), ctx)
			if err != nil {
				log.Error("Read error on glob: ", err)
			} else {
				var curFile *types.File
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

func (b *B2Backend) RunGC(ctx context.Context) ([]types.File, error) {
	log := logger.LogFromCtx(packageName+".RunGC", ctx)
	var deletedFiles = []types.File{}

	log.Info("Obtaining file list from backend")
	fPtrs, err := b.ListGlob(ctx, "*")
	if err != nil {
		return nil, err
	}

	log.Info("Scanning for files to be deleted")
	for _, v := range fPtrs {
		if v.DeleteAt.After(time.Now().UTC()) {
			log.Debug("Scheduling ", v.Flake, " for deletion")
			deletedFiles = append(deletedFiles, *v)
		}
	}

	// Deletion is put into a second step to A) speed up scan and B)
	// be more resilient (we can return a full list of maybe GC'd data)
	//
	// Making a second run and comparing the returned lists may reveal
	// some error points.
	log.Info("Deleting files")
	for _, v := range fPtrs {
		err := b.Delete(v.Flake, ctx)
		if err != nil {
			log.Debug("Deleting flake ", v.Flake)
			return deletedFiles, err
		}
	}

	return deletedFiles, nil
}
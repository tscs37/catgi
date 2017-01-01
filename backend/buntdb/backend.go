package buntdb

import (
	"encoding/json"

	"context"

	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/Sirupsen/logrus"
	"github.com/tidwall/buntdb"
)

const bePackagename = "backend/buntdb/backend"

type BuntDBBackend struct {
	db *buntdb.DB
}

func (b *BuntDBBackend) Name() string { return "buntdb-backend" }

func (b *BuntDBBackend) Upload(name string, file *types.File, ctx context.Context) error {
	log := logger.LogFromCtx(bePackagename+".Upload", ctx)
	return b.db.Update(func(tx *buntdb.Tx) error {
		log.Debug("Storing file ", name)
		log.Debug("Checking DB, should return not found")
		_, err := tx.Get("/file/" + name)
		if err != buntdb.ErrNotFound {
			log.Debug("Error was not ErrNotFound")
			return err
		}
		log.Debug("Encode File to JSON")
		encoded, err := json.Marshal(file)
		if err != nil {
			logrus.Debug("Encoding Error ", err)
			return err
		}
		log.Debug("Storing JSON into DB")
		_, _, err = tx.Set("/file/"+name, string(encoded), nil)
		return err
	})
}

func (b *BuntDBBackend) Exists(name string, ctx context.Context) error {
	return b.db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get("/file/" + name)
		if err != buntdb.ErrNotFound {
			return err
		}
		return nil
	})
}

func (b *BuntDBBackend) Get(name string, ctx context.Context) (*types.File, error) {
	var file = &types.File{}
	log := logger.LogFromCtx(bePackagename+".Get", ctx)

	errTx := b.db.View(func(tx *buntdb.Tx) error {
		log.Debug("Getting file ", name)
		dat, err := tx.Get("/file/" + name)
		if err != nil {
			log.Debug("File does not exist, returning error from Tx")
			return err
		}
		log.Debug("Unmarshalling file")
		return json.Unmarshal([]byte(dat), file)
	})

	return file, errTx
}

func (b *BuntDBBackend) Delete(name string, ctx context.Context) error {
	return b.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete("/file/" + name)
		return err
	})
}

func (b *BuntDBBackend) ListGlob(ctx context.Context, prefix string) ([]*types.File, error) {
	log := logger.LogFromCtx(bePackagename+".ListGlob", ctx)
	files := make([]*types.File, 0)
	b.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys("/file/"+prefix+"*", func(key, value string) bool {
			var next *types.File
			err := json.Unmarshal([]byte(value), next)
			if err != nil {
				log.Error("Error during Bunt KV: ", err)
			} else {
				files = append(files, next)
			}
			return true
		})
	})
	return nil, types.ErrorNotImplemented
}

func (b *BuntDBBackend) Publish(_ []string, _ string, _ context.Context) error {
	return types.ErrorNotImplemented
}

func (b *BuntDBBackend) Unpublish(_ string, _ context.Context) error {
	return types.ErrorNotImplemented
}

func (b *BuntDBBackend) Resolve(_ string, _ context.Context) ([]string, error) {
	return nil, types.ErrorNotImplemented
}

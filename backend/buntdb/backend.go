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

func (b *BuntDBBackend) ListGlob(
	glob string, ictx context.Context) (
	[]*types.File, context.Context, error) {
	return nil, nil, nil
}

func (b *BuntDBBackend) CleanUp(ctx context.Context) error {
	return nil
}

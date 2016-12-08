package buntdb

import (
	"encoding/json"

	"git.timschuster.info/rls.moe/catgi/backend/types"
	"github.com/Sirupsen/logrus"
	"github.com/tidwall/buntdb"
)

type BuntDBBackend struct {
	db *buntdb.DB
}

func (b *BuntDBBackend) Name() string { return "buntdb-backend" }

func (b *BuntDBBackend) Upload(name string, file *types.File) error {
	return b.db.Update(func(tx *buntdb.Tx) error {
		logrus.Debug("Storing file ", name)
		logrus.Debug("Checking DB, should return not found")
		_, err := tx.Get("/file/" + name)
		if err != buntdb.ErrNotFound {
			logrus.Debug("Error was not ErrNotFound")
			return err
		}
		logrus.Debug("Encode File to JSON")
		encoded, err := json.Marshal(file)
		if err != nil {
			logrus.Debug("Encoding Error ", err)
			return err
		}
		logrus.Debug("Storing JSON into DB")
		_, _, err = tx.Set("/file/"+name, string(encoded), nil)
		return err
	})
}

func (b *BuntDBBackend) Exists(name string) error {
	return b.db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get("/file/" + name)
		if err != buntdb.ErrNotFound {
			return err
		}
		return nil
	})
}

func (b *BuntDBBackend) Get(name string) (*types.File, error) {
	var file = &types.File{}

	errTx := b.db.View(func(tx *buntdb.Tx) error {
		logrus.Debug("Getting file ", name)
		dat, err := tx.Get("/file/" + name)
		if err != nil {
			logrus.Debug("File does not exist, returning error from Tx")
			return err
		}
		logrus.Debug("Unmarshalling file")
		return json.Unmarshal([]byte(dat), file)
	})

	return file, errTx
}

func (b *BuntDBBackend) Delete(name string) error {
	return b.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete("/file/" + name)
		return err
	})
}

func (b *BuntDBBackend) LoadIndex(i types.Index) error {
	return b.db.View(func(tx *buntdb.Tx) error {
		dat, err := tx.Get("/index/")
		if err != nil {
			return err
		}
		return i.Unserialize([]byte(dat))
	})
}

func (b *BuntDBBackend) StoreIndex(i types.Index) error {
	return b.db.Update(func(tx *buntdb.Tx) error {
		dat, err := i.Serialize()
		if err != nil {
			return err
		}
		_, _, err = tx.Set("/index/", string(dat), &buntdb.SetOptions{
			Expires: false,
		})
		return err
	})
}

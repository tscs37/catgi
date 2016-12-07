package buntdb

import (
	"encoding/json"
	"time"

	"git.timschuster.info/rls.moe/catgi/backend/types"
	"github.com/tidwall/buntdb"
)

type BuntDBBackend struct {
	db *buntdb.DB
}

var defKeyOpts = buntdb.SetOptions{
	Expires: true,
	TTL:     time.Hour * 24 * 30,
}

func (b *BuntDBBackend) Name() string { return "buntdb-backend" }

func (b *BuntDBBackend) Upload(name string, ttl *time.Duration, file *types.File) error {
	return b.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Get("/file/" + name)
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(file)
		if err != nil {
			return err
		}
		keyopts := defKeyOpts
		if ttl != nil {
			keyopts.TTL = *ttl
		}
		_, _, err = tx.Set("/file/"+name, string(encoded), &keyopts)
		return err
	})
}

func (b *BuntDBBackend) Exists(name string) error {
	return b.db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get("/file/" + name)
		if err != nil {
			return err
		}
		return nil
	})
}

func (b *BuntDBBackend) Get(name string) (*types.File, error) {
	var file *types.File

	errTx := b.db.View(func(tx *buntdb.Tx) error {
		dat, err := tx.Get("/file/" + name)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(dat), file)
	})

	return file, errTx
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

func (b *BuntDBBackend) Delete(name string) error {
	return b.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete("/file/" + name)
		return err
	})
}

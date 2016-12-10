package buntdb

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"git.timschuster.info/rls.moe/catgi/backend/types"

	"strconv"

	"git.timschuster.info/rls.moe/catgi/snowflakes"
	"github.com/Sirupsen/logrus"
	"github.com/tidwall/buntdb"
)

type BuntDBIndex struct {
	db *buntdb.DB
}

func (i *BuntDBIndex) Name() string {
	return "buntdb-index"
}

func (i *BuntDBIndex) Serialize(ctx context.Context) ([]byte, error) {
	var buf = bytes.NewBuffer([]byte{})
	if err := i.db.Save(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (i *BuntDBIndex) Unserialize(s []byte, ctx context.Context) error {
	return i.db.Load(bytes.NewReader(s))
}

func (i *BuntDBIndex) Get(r types.File, b types.Backend, ctx context.Context) (bool, *types.File, error) {
	err := i.checkAndClearFlake(r.Flake, b, ctx)
	if err != nil {
		return false, nil, err
	}

	file, err := b.Get(r.Flake, ctx)

	if file != nil {
		file.Flake = r.Flake
	}

	return false, file, err
}

func (i *BuntDBIndex) Put(r types.File, b types.Backend, ctx context.Context) (bool, *types.File, error) {
	var file = &types.File{}

	// Safely Copy the struct data to the file
	file.CreatedAt = time.Now().UTC()
	file.Data = r.Data
	file.Public = r.Public

	file.DeleteAt = r.DeleteAt

	logrus.Debug("Determining File.DeleteAt...")
	if file.DeleteAt.Unix() == 0 {
		logrus.Debug("Using default TTL")
		file.DeleteAt = time.Now().UTC().Add(types.DefaultTTL)
	} else if file.DeleteAt.Sub(time.Now().UTC()) > types.MaxTTL {
		logrus.Debug("Using max TTL")
		file.DeleteAt = time.Now().UTC().Add(types.MaxTTL)
	} else if file.DeleteAt.Sub(time.Now().UTC()) < types.MinTTL {
		logrus.Debug("Using min TTL")
		file.DeleteAt = time.Now().UTC().Add(types.MinTTL)
	} else {
		logrus.Debug("TTL: %fh", file.DeleteAt.Sub(time.Now().UTC()).Hours())
	}

	flake, err := snowflakes.NewSnowflake()
	if err != nil {
		return false, nil, err
	}
	file.Flake = flake

	err = i.db.Update(func(tx *buntdb.Tx) error {
		err = b.Upload(flake, file, ctx)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("/meta/file/"+flake+"/delete_at/", fmt.Sprintf("%d", file.DeleteAt.Unix()), nil)
		if err != nil {
			tx.Rollback()
			return err
		}
		if file.Public {
			_, _, err = tx.Set("/meta/file/"+flake+"/public/", fmt.Sprintf("%t", file.Public), nil)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
		_, _, err = tx.Set("/meta/file/"+flake+"/created_at/",
			fmt.Sprintf("%d", file.CreatedAt.Unix()), nil)
		if err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})
	if err != nil {
		return false, nil, err

	}

	return false, file, err
}

func (i *BuntDBIndex) Del(r types.File, b types.Backend, ctx context.Context) (bool, error) {
	return false, b.Delete(r.Flake, ctx)
}

// Flush and Clear do nothing on this simple Index yet

func (i *BuntDBIndex) Flush(_ context.Context) error { return nil }
func (i *BuntDBIndex) Clear(_ context.Context) error { return nil }

// checkAndClearFlake deletes flakes that are expired
func (i *BuntDBIndex) checkAndClearFlake(
	flake string, be types.Backend, ctx context.Context) error {

	return i.db.Update(func(tx *buntdb.Tx) error {
		logrus.Debug("Checking if file meta exists")
		dat, err := tx.Get("/meta/file/" + flake + "/delete_at/")
		if err == buntdb.ErrNotFound {
			logrus.Debug("File meta not found, skipping")
			return nil
		} else if err != nil {
			logrus.Debug("File meta error, aborting")
			return err
		}
		logrus.Debug("Found, parsing Expiry Date")
		unixDelAt, err := strconv.ParseInt(dat, 10, 64)
		if err != nil {
			logrus.Debug("Error on convert: %s", err)
			return err
		}
		if !time.Unix(unixDelAt, 0).After(time.Now().UTC()) {
			logrus.Debug("File expired, deleting")
			err := be.Delete(flake, ctx)
			if err != nil {
				logrus.Debug("Deletion on backend failed, aborting")
				tx.Rollback()
				return err
			}
			_, err = tx.Delete("/meta/file/" + flake + "/delete_at/")
			if err != nil {
				logrus.Debug("Deletion on index failed: deleted_at: ", err)
				tx.Rollback()
				return err
			}
			_, err = tx.Delete("/meta/file/" + flake + "/public/")
			if err != nil {
				logrus.Debug("Deletion on index failed: public: ", err)
				tx.Rollback()
				return err
			}
			_, err = tx.Delete("/meta/file/" + flake + "/created_at/")
			if err != nil {
				logrus.Debug("Deletion on index failed: created_at: ", err)
				tx.Rollback()
				return err
			}
		} else {
			logrus.Debugf("File expires in %f minutes", time.Unix(unixDelAt, 0).Sub(time.Now().UTC()).Minutes())
		}
		logrus.Debug("Clean finished. Returning.")
		return nil
	})
}

func (i *BuntDBIndex) Collect(be types.Backend, ctx context.Context) error {
	return i.db.Update(func(tx *buntdb.Tx) error {
		//tx.AscendLessThan("index-file-deleteat")
		return nil
	})
}

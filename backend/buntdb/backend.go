package buntdb

import (
	"encoding/json"

	"context"

	"os"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/Sirupsen/logrus"
	"github.com/tidwall/buntdb"
)

const bePackagename = "backend/buntdb/backend"

type BuntDBBackend struct {
	db      *buntdb.DB
	file    string
	autoTTL bool
}

func (b *BuntDBBackend) Name() string { return "buntdb-backend" }

func (b *BuntDBBackend) Upload(name string, file *common.File, ctx context.Context) error {
	log := logger.LogFromCtx(bePackagename+".Upload", ctx)

	if file == nil {
		return common.ErrorSerializationFailure
	}

	if file.Flake != name {
		log.Debug("Flake mismatch, correcting flake in file")
		file.Flake = name
	}

	return b.db.Update(func(tx *buntdb.Tx) error {
		log.Debug("Storing file ", name)

		log.Debug("Checking DB, should return not found")
		_, err := tx.Get("/file/" + name)
		if err != buntdb.ErrNotFound {
			tx.Rollback()
			log.Debug("Error was not ErrNotFound")
			return err
		}

		// BuntDB stores the entire file, separating the data
		// and metadata is not necessary
		log.Debug("Encode File to JSON")
		encoded, err := json.Marshal(file)
		if err != nil {
			tx.Rollback()
			logrus.Debug("Encoding Error ", err)
			return err
		}

		var opts *buntdb.SetOptions = nil
		if file.DeleteAt != nil && b.autoTTL {
			opts = &buntdb.SetOptions{
				Expires: true,
				TTL:     file.DeleteAt.TTL(),
			}
		}

		log.Debug("Storing JSON into DB")
		_, _, err = tx.Set("/file/"+name, string(encoded), opts)

		if err != nil {
			tx.Rollback()
			log.Error("Error in tx: ", err)
			return err
		}
		return nil
	})
}

func (b *BuntDBBackend) Exists(name string, ctx context.Context) error {
	return b.db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get("/file/" + name)
		if err != buntdb.ErrNotFound {
			return err
		} else if err == buntdb.ErrNotFound {
			return common.NewErrorFileNotExists(name, err)
		}
		return nil
	})
}

func (b *BuntDBBackend) Get(name string, ctx context.Context) (*common.File, error) {
	var file = &common.File{}
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

	if file.Data == nil {
		file.Data = []byte{}
	}

	return file, errTx
}

func (b *BuntDBBackend) Delete(name string, ctx context.Context) error {
	return b.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete("/file/" + name)
		return err
	})
}

func (b *BuntDBBackend) ListGlob(ctx context.Context, prefix string) ([]*common.File, error) {
	log := logger.LogFromCtx(bePackagename+".ListGlob", ctx)
	files := make([]*common.File, 0)
	b.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys("/file/"+prefix, func(key, value string) bool {
			var next = &common.File{}
			err := json.Unmarshal([]byte(value), next)
			if err != nil {
				log.Error("Error during Bunt KV: ", err)
				log.Debug("Data Value: '%s'", value)
			} else {
				files = append(files, next)
			}
			return true
		})
	})
	return files, nil
}

// RunGC will try to find expired files, usually Bunt will take care of
// this but this should cleanup any orphaned entries.
// TODO: Remove once automatic expiry is properly tested
func (b *BuntDBBackend) RunGC(ctx context.Context) ([]common.File, error) {
	log := logger.LogFromCtx(bePackagename+".RunGC", ctx)
	var deletedFiles = []common.File{}

	log.Debug("Obtaining file list from backend")
	fPtrs, err := b.ListGlob(ctx, "*")
	if err != nil {
		log.Error("Error on Obtaining List: ", err)
		return nil, err
	}

	log.Debugf("About to clean %d files", len(fPtrs))

	log.Debug("Scanning for files to be deleted")
	for _, v := range fPtrs {
		if v.DeleteAt == nil {
			log.Warn("File contains NIL DeleteAt: ", v.Flake)
			continue
		}
		if v.DeleteAt.TTL() <= 0 {
			log.Debug("Scheduling ", v.Flake, " for deletion")
			v.Data = []byte{}
			deletedFiles = append(deletedFiles, *v)
		}
	}

	// Deletion is put into a second step to A) speed up scan and B)
	// be more resilient (we can return a full list of maybe GC'd data)
	//
	// Making a second run and comparing the returned lists may reveal
	// some error points.
	log.Debugf("Deleting files")
	for _, v := range deletedFiles {
		log.Debug("Starting delete for ", v.Flake)
		if common.IsFileNotExists(b.Exists(v.Flake, ctx)) {
			log.Debug("Flake already expired, skipping")
			continue
		}
		err := b.Delete(v.Flake, ctx)
		if err != nil {
			log.Debug("Error while deleting flake ", v.Flake)
			return deletedFiles, err
		}
	}

	log.Debugf("Deleted %d flakes", len(deletedFiles))

	log.Debugf("Shrunk DB by %d bytes", b.shrink(ctx))

	return deletedFiles, nil
}

// Shrink attempts to reduce the size of the DB file and returns
// the number of bytes saved.
// If an error occurs, it's ignored.
// If the DB is running in memory mode, it returns 0 as cleaned size
// as memory cannot be reliably stat'd.
func (b *BuntDBBackend) shrink(ctx context.Context) int64 {
	log := logger.LogFromCtx(bePackagename+".shrink", ctx)

	var startSize, endSize int64

	if b.file != ":memory:" {
		log.Debugf("Stat'ing file %s", b.file)
		if stat, err := os.Stat(b.file); !os.IsNotExist(err) {
			startSize = stat.Size()
			log.Debugf("Begin Shrink with %d bytes", startSize)
		} else {
			log.Error("Error while stat'ing DB: ", err)
		}
	}

	log.Debug("Beginning Shrink")
	err := b.db.Shrink()
	if err != nil {
		log.WithField("non-critical", "").Warn("DB shrink failed: ", err)
	} else {
		log.Debug("Shrink OK")
	}

	if b.file != ":memory:" {
		log.Debugf("Stat'ing file %s", b.file)
		if stat, err := os.Stat(b.file); !os.IsNotExist(err) {
			endSize = stat.Size()
			log.Debugf("Ended Shrink with %d bytes", endSize)
		} else {
			log.Error("Error while stat'ing DB: ", err)
		}
	}

	shrink := startSize - endSize

	log.Debugf("Result shrink of %d kib", (shrink / 1024))
	return shrink
}

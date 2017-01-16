package common

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/logger"
)

// GenericGC is a generic garbage collector implementation
// that allows backends to utilize already existing methods
// for GC instead of having to write their own.
//
// If your backend requires some pre-exec or post-exec work
// you can supply this in the appropriate function parameter.
// An error in the PreHook will abort the function while an
// error in the PostHook is report-only.
//
// The Hooks will only run if they are non-nil, if they are nil
// they are ignored.
func GenericGC(b Backend,
	preHook, postHook func(Backend, logger.Logger) error,
	ctx context.Context) ([]File, error) {

	log := logger.LogFromCtx("GenericGC."+b.Name()+".RunGC", ctx)

	if preHook != nil {
		log.Debug("Running PreHook")

		err := preHook(b, log)
		if err != nil {
			log.Error("Error in GC PreHook")
			return nil, err
		}
	}

	var deletedFiles = []File{}

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
		if IsFileNotExists(b.Exists(v.Flake, ctx)) {
			log.Debug("Flake already expired, skipping")
			continue
		}
		err = b.Delete(v.Flake, ctx)
		if err != nil {
			log.Debug("Error while deleting flake ", v.Flake)
			return deletedFiles, err
		}
	}

	log.Debugf("Deleted %d flakes", len(deletedFiles))

	if postHook != nil {
		log.Debug("Running PostHook")

		err = postHook(b, log)
		if err != nil {
			log.Error("Error in GC PostHook")
			return deletedFiles, err
		}
	}

	return deletedFiles, nil
}

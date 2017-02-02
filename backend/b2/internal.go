package b2

import (
	"bytes"
	"context"
	"io"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/kurin/blazer/b2"
)

// writeFile writes raw data into a specified file and logs into a context
func (b *B2Backend) writeFile(name string, data []byte, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".writeFile", ctx).
		WithField("object", name).WithField("obj_len", len(data))
	if len(data) > common.MaxDataSize {
		log.Warn("Attempted to store way too large file.")
		return common.ErrorQuotaExceeded
	}
	obj := b.dataBucket.Object(name)
	log.Debug("Opening new Writer")
	w := obj.NewWriter(ctx)
	defer w.Close()
	log.Debug("Creating new Data Buffer")
	buf := bytes.NewBuffer(data)
	n, err := io.Copy(w, buf)
	if err != nil {
		log.Error("Error while uploading: ", err)
		return err
	}
	log.Debugf("Wrote %d bytes", n)
	return nil
}

// deleteFile is a wrapper around b2.Bucket.Object().Delete()
func (b *B2Backend) deleteFile(name string, ctx context.Context) error {
	return b.dataBucket.Object(name).Delete(ctx)
}

// pingFile returns a boolean indicating wether a file exists and additionally
// it's attributes. Otherwise it returns false and an error.
// The attributes returned may not be null and contain data even if
// the boolean is false. For example when a file is currently being
// uploaded.
func (b *B2Backend) pingFile(name string, ctx context.Context) (bool, *b2.Attrs, error) {
	log := logger.LogFromCtx(packageName+".pingFile", ctx).
		WithField("object", name)
	obj := b.dataBucket.Object(name)
	log.Debug("Loading object attributes")
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return false, nil, common.NewErrorFileNotExists(name, err)
	}

	if attrs.Status == b2.Uploaded {
		return true, attrs, nil
	}
	return false, attrs, nil
}

// readFile returns the entire contents of the file or an error.
func (b *B2Backend) readFile(name string, ctx context.Context) ([]byte, error) {
	log := logger.LogFromCtx(packageName+".readFile", ctx).
		WithField("object", name)
	obj := b.dataBucket.Object(name)
	log.Debug("Loading object attributes")

	r := obj.NewReader(ctx)

	buffer := bytes.NewBuffer([]byte{})
	n, err := io.Copy(buffer, r)
	if err != nil {
		log.Error("Error while reading: ", err)
		return nil, err
	}
	log.Debugf("Read %d bytes", n)
	return buffer.Bytes(), nil
}

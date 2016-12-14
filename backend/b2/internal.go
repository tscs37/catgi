package backend

import (
	"bytes"
	"context"
	"io"
	"strings"

	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/kurin/blazer/b2"
)

func dataName(flake string) string { return "file/" + splitName(flake) + "/data.bin" }

func metaName(flake string) string { return "file/" + splitName(flake) + "/meta.json" }

func pubName(flake string) string { return "public/" + flake }

func clpubName(name string) string { return "named/" + name + "/flakes.json" }

func (b *B2Backend) writeFile(name string, data []byte, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".writeFile", ctx).
		WithField("object", name).WithField("obj_len", len(data))
	obj := b.dataBucket.Object(name)
	log.Debug("Opening new Writer")
	w := obj.NewWriter(ctx)
	defer w.Close()
	log.Debug("Creating new Data Buffer")
	buf := bytes.NewBuffer(data)
	if n, err := io.Copy(w, buf); err != nil {
		log.Error("Error while uploading: ", err)
		return err
	} else {
		log.Debugf("Wrote %d bytes", n)
		return nil
	}
}

func (b *B2Backend) deleteFile(name string, ctx context.Context) error {
	return b.dataBucket.Object(name).Delete(ctx)
}

func (b *B2Backend) pingFile(name string, ctx context.Context) (bool, *b2.Attrs, error) {
	log := logger.LogFromCtx(packageName+".pingFile", ctx).
		WithField("object", name)
	obj := b.dataBucket.Object(name)
	log.Debug("Loading object attributes")
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return false, nil, types.NewErrorFileNotExists(name, err)
	}

	if attrs.Status == b2.Uploaded {
		return true, attrs, nil
	}
	return false, attrs, nil
}

func (b *B2Backend) readFile(name string, ctx context.Context) ([]byte, error) {
	log := logger.LogFromCtx(packageName+".readFile", ctx).
		WithField("object", name)
	obj := b.dataBucket.Object(name)
	log.Debug("Loading object attributes")

	r := obj.NewReader(ctx)

	buffer := bytes.NewBuffer([]byte{})
	if n, err := io.Copy(buffer, r); err != nil {
		log.Error("Error while reading: ", err)
		return nil, err
	} else {
		log.Debugf("Read %d bytes", n)
		return buffer.Bytes(), nil
	}
}

func splitName(flakeStr string) string {
	flake := []rune(flakeStr)
	var out = []string{}
	skipSize := 2
	for true {
		if len(flake) > skipSize+1 {
			out = append(out, string(flake[0:skipSize]))
			flake = flake[skipSize:]
		} else {
			out = append(out, string(flake))
			return strings.Join(out, "/")
		}
	}
	panic("Should not terminate here")
}

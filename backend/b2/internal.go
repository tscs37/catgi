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

// dataName is used to store the raw binary data for a file
// Format: "file/<flake>/public.json"
func dataName(flake string) string { return "file/" + splitName(flake) + "/data.bin" }

// metaName is used to store metainformation for a file
// Format: "file/<flake>/meta.json"
func metaName(flake string) string { return "file/" + splitName(flake) + "/meta.json" }

func isMetaFile(file string) bool {
	return strings.HasPrefix(file, "file/") && strings.HasSuffix(file, "/meta.json")
}

func isDataFile(file string) bool {
	return strings.HasPrefix(file, "file/") && strings.HasSuffix(file, "/data.bin")
}

// pubName is used to store public flakes for iteration
// Format: "public/<flake>"
func pubName(flake string) string { return "public/" + flake }

// clpubName is used to store published names
// Format: "named/<name>/flakes.json"
func clpubName(name string) string { return "named/" + name + "/flakes.json" }

// writeFile writes raw data into a specified file and logs into a context
func (b *B2Backend) writeFile(name string, data []byte, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".writeFile", ctx).
		WithField("object", name).WithField("obj_len", len(data))
	if len(data) > types.MaxDataSize {
		log.Warn("Attempted to store way too large file.")
		return types.ErrorQuotaExceeded
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
		return false, nil, types.NewErrorFileNotExists(name, err)
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

// splitName splits a string according to the following rules:
// 1. Create a slice of strings
// 2. If the remaining size of the string is larger than skipSize plus 1
//      then take the first 2 runes and append them as string to the slice
// 3. If this is not the case, take all remaining runes and append
//      them to the slice
// 4. Join all slice elements with "/" inbetween.
//
// skipSize is by default 2
//
// Example:
//
// HelloWorld       =>      He/ll/oW/or/ld
// HelloInternet    =>      He/ll/oI/nt/er/net
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

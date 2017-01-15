package compltest

import (
	"testing"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"github.com/stretchr/testify/assert"
)

func testUploadNil(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	err := b.Upload(nilTest, nil, ctx)

	assert.Error(err, "Must not be able to upload nil")
}

func testUploadEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	empty := &common.File{}
	err := b.Upload(emptyTest, empty, ctx)

	assert.NoError(err, "Empty files must be uploadable")
}

func testUploadNonEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	file := &common.File{}
	file.ContentType = "text/html"
	file.CreatedAt = common.FromTime(packageTime)
	file.Data = []byte("<html>test</html>")
	file.DeleteAt = common.FromTime(packageTime.AddDate(100, 11, 200))
	file.FileExtension = ".html"
	file.Flake = "index.html"
	file.Public = false
	file.User = "testuser"

	err := b.Upload(nonEmptyTest, file, ctx)

	assert.NoError(err, "Uploading non-empty file must not return error")
}

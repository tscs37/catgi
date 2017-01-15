package compltest

import (
	"context"
	"testing"
	"time"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/stretchr/testify/assert"
)

// RunTestSuite will run a test suite over the Backend
// to ensure it properly functions as defined
func RunTestSuite(b common.Backend, t *testing.T) {
	testGetNonExist(b, t)
	testUploadEmpty(b, t)
	testUploadNonEmpty(b, t)
	testGetEmpty(b, t)
}

// GetTestCtx returns a context suitable for test usage
func GetTestCtx() context.Context {
	ctx := context.Background()
	ctx = logger.InjectLogToContext(ctx)
	return ctx
}

func testRegistration(b common.Backend, t *testing.T) {
	for _, v := range backend.InstalledDrivers() {
		if v == b.Name() {
			return
		}
	}
	t.Error("Driver did not register on import")
	t.Fail()
}

func testGetNonExist(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	err := b.Exists("non-exist", ctx)

	assert.Error(err, "Non existant file must return error")
	assert.True(common.IsFileNotExists(err), "Error must recognized as ErrorFileNotExist error")
}

func testUploadEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	empty := &common.File{}
	err := b.Upload("empty-test-file", empty, ctx)

	assert.NoError(err, "Empty files must be uploadable")
}

func testUploadNonEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	file := &common.File{}
	file.ContentType = "text/html"
	file.CreatedAt = common.FromTime(time.Now())
	file.Data = []byte("<html>test</html>")
	file.DeleteAt = common.FromTime(time.Now().AddDate(100, 11, 200))
	file.FileExtension = ".html"
	file.Flake = "index.html"
	file.Public = false
	file.User = "testuser"

	err := b.Upload("index-test-file", file, ctx)

	assert.NoError(err, "Uploading non-empty file must not return error")
}

func testGetEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)
	file, err := b.Get("empty-test-file", ctx)

	assert.NoError(err, "Must be able to retrieve empty file")

	assert.Nil(file.CreatedAt, "CreatedAt must be nil on empty file")
	assert.Nil(file.DeleteAt, "DeleteAt must be nil on empty file")

	assert.NotNil(file.Data, "File Data may never be nil on empty file")
	assert.Empty(file.Data, "File Data must be empty on empty file")

	assert.Empty(file.ContentType, "Content Type must be empty on empty file")
	assert.NotNil(file.ContentType, "Content Type must not be nil on empty file")

	assert.Empty(file.FileExtension, "File Extension must be empty on empty file")
	assert.NotNil(file.FileExtension, "File Extension must not be nil on emtpy file")

	// The flake should not be nil and since it was empty it should have been
	// corrected to the flake used when storing the file
	assert.NotNil(file.Flake, "Flake must not be nil on empty file")
	assert.EqualValues(file.Flake, "empty-test-file", "Flake must be corrected on empty file")

	assert.False(file.Public, "Public must be false on empty file")

	assert.Empty(file.User, "User must be empty on empty file")
	assert.NotNil(file.User, "User must not be nil on empty file")
}

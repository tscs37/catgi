package compltest

import (
	"testing"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"github.com/stretchr/testify/assert"
)

func testGetEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)
	file, err := b.Get(emptyTest, ctx)

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

func testGetNonEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	file := &common.File{}
	file.ContentType = "text/html"
	file.CreatedAt = common.FromTime(packageTime)
	file.Data = []byte("<html>test</html>")
	file.DeleteAt = common.FromTime(packageTime.AddDate(100, 11, 200))
	file.FileExtension = ".html"
	// file.Flake = "index.html"
	// The backend must amend this
	file.Flake = nonEmptyTest
	file.Public = false
	file.User = "testuser"

	backendFile, err := b.Get(nonEmptyTest, ctx)

	assert.EqualValues(file, backendFile, "Backend file must match stored file")
	assert.NoError(err, "Must not return error")
}

func testGetNonExist(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	err := b.Exists(notExist, ctx)

	assert.Error(err, "Non existant file must return error")
	assert.True(common.IsFileNotExists(err), "Error must recognized as ErrorFileNotExist error")
}

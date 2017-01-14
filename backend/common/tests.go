package common

import (
	"context"
	"testing"
	"time"

	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/stretchr/testify/assert"
)

// RunTestSuite will run a test suite over the Backend
// to ensure it properly functions as defined
func RunTestSuite(b Backend, t *testing.T) {
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
func testGetNonExist(b Backend, t *testing.T) {
	ctx := GetTestCtx()

	err := b.Exists("non-exist", ctx)
	if !IsFileNotExists(err) {
		t.Error("b.Exists did not return error on non-existant file: ", err)
		t.Fail()
		return
	}
}

func testUploadEmpty(b Backend, t *testing.T) {
	ctx := GetTestCtx()
	empty := &File{}
	err := b.Upload("empty-test-file", empty, ctx)
	if err != nil {
		t.Error("Error on upload: ", err)
		t.Fail()
		return
	}
}

func testUploadNonEmpty(b Backend, t *testing.T) {
	ctx := GetTestCtx()
	file := &File{}
	file.ContentType = "text/html"
	file.CreatedAt = FromTime(time.Now())
	file.Data = []byte("<html>test</html>")
	file.DeleteAt = FromTime(time.Now().AddDate(100, 11, 200))
	file.FileExtension = ".html"
	file.Flake = "index.html"
	file.Public = false
	file.User = "testuser"

	err := b.Upload("index-test-file", file, ctx)
	if err != nil {
		t.Error("Error on upload: ", err)
		t.Fail()
		return
	}
}

func testGetEmpty(b Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)
	file, err := b.Get("empty-test-file", ctx)
	if err != nil {
		t.Error("Error on empty retrieve: ", err)
		t.Fail()
		return
	}

	assert.Nil(file.CreatedAt)
	assert.Nil(file.DeleteAt)

	// File Data must never be nil
	assert.NotNil(file.Data)
	assert.Empty(file.Data)

	assert.Empty(file.ContentType)
	assert.NotNil(file.ContentType)

	assert.Empty(file.FileExtension)
	assert.NotNil(file.FileExtension)

	assert.Empty(file.Flake)
	assert.NotNil(file.Flake)

	assert.False(file.Public)

	assert.Empty(file.User)
	assert.NotNil(file.User)
}

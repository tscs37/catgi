package compltest

import (
	"testing"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"github.com/stretchr/testify/assert"
)

func testExistNoExist(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	err := b.Exists(notExist, ctx)

	assert.Error(err, "Must return error")
	assert.True(common.IsFileNotExists(err), "Must be FileNotExists error")
}

func testExistEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	err := b.Exists(emptyTest, ctx)

	assert.NoError(err, "Must not return error on empty file")
}

func testExistNonEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	err := b.Exists(nonEmptyTest, ctx)

	assert.NoError(err, "Must not return error on empty file")
}

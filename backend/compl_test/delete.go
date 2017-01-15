package compltest

import (
	"testing"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"github.com/stretchr/testify/assert"
)

func testDeleteNonEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	err := b.Delete(nonEmptyTest, ctx)

	assert.NoError(err, "Must be able to delete empty file")
}

func testDeleteEmpty(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	err := b.Delete(emptyTest, ctx)

	assert.NoError(err, "Must be able to delete empty file")
}

func testDeleteNoExist(b common.Backend, t *testing.T) {
	ctx := GetTestCtx()
	assert := assert.New(t)

	err := b.Delete("does-not-exist", ctx)

	assert.Error(err, "Must not be able to delete empty file")
}

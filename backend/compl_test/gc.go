package compltest

import (
	"testing"
	"time"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"github.com/stretchr/testify/assert"
)

func testGC(b common.Backend, t *testing.T) {
	assert := assert.New(t)
	ctx := GetTestCtx()

	gcFile := &common.File{}
	gcFile.Data = []byte{}
	gcFile.DeleteAt = common.FromTime(time.Now().AddDate(-1, 0, 0))
	noGcFile := &common.File{}
	noGcFile.Data = []byte{}
	noGcFile.DeleteAt = common.FromTime(time.Now().AddDate(1, 0, 0))

	err := b.Upload(gcTest, gcFile, ctx)

	assert.NoError(err, "Must not return error")

	err = b.Upload(noGcTest, noGcFile, ctx)

	assert.NoError(err, "Must not return error")

	f, err := b.RunGC(ctx)

	assert.NoError(err, "Must not return error from GC")
	assert.True(len(f) == 0 || len(f) == 1, "Must return either nothing or 1 file")
	if len(f) == 1 {
		assert.EqualValues(*gcFile, f[0], "Return must contain the file that was gc'd")
	}
}

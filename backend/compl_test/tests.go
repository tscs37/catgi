package compltest

import (
	"context"
	"testing"
	"time"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
)

var (
	packageTime time.Time
)

func init() {
	var err error
	packageTime, err = time.Parse("2006-01-02", "2017-12-10")
	if err != nil {
		panic(err)
	}
}

const (
	nonEmptyTest = "index-test-file"
	emptyTest    = "empty-test-file"
	notExist     = "does-not-exist"
	nilTest      = "nil-test"
	gcTest       = "gc-test"
	noGcTest     = "no-gc-test"
)

// RunTestSuite will run a test suite over the Backend
// to ensure it properly functions as defined
func RunTestSuite(b common.Backend, t *testing.T) {
	testUploadNil(b, t)
	testUploadEmpty(b, t)
	testUploadNonEmpty(b, t)

	testGetEmpty(b, t)
	testGetNonEmpty(b, t)
	testGetNonExist(b, t)

	testExistNoExist(b, t)
	testExistEmpty(b, t)
	testExistNonEmpty(b, t)

	testGC(b, t)

	testDeleteEmpty(b, t)
	testDeleteNoExist(b, t)
	testDeleteNonEmpty(b, t)
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

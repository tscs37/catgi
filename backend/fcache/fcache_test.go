package fcache

import (
	"testing"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/backend/compl_test"

	_ "git.timschuster.info/rls.moe/catgi/backend/buntdb"
)

func getTestDB() (common.Backend, error) {
	return NewFCacheBackend(map[string]interface{}{
		"driver": "buntdb",
		"params": map[string]interface{}{
			"file": ":memory:",
		},
		"cache_size": 10,
	}, compltest.GetTestCtx())
}

func TestCompliance(t *testing.T) {
	fcacheWithBuntdb, err := getTestDB()

	if err != nil {
		t.Log("Error on creating Testing Backend: ", err)
		t.FailNow()
		return
	}

	compltest.RunTestSuite(fcacheWithBuntdb, t)
}

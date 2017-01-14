package fcache

import (
	"testing"

	"git.timschuster.info/rls.moe/catgi/backend/common"

	_ "git.timschuster.info/rls.moe/catgi/backend/buntdb"
)

func TestCompliance(t *testing.T) {
	ctx := common.GetTestCtx()
	fcacheWithBuntdb, err := NewFCacheBackend(map[string]interface{}{
		"driver": "buntdb",
		"params": map[string]interface{}{
			"file": ":memory:",
		},
		"cache_size": 10,
	}, ctx)

	if err != nil {
		t.Log("Error on creating Testing Backend: ", err)
		t.FailNow()
		return
	}

	common.RunTestSuite(fcacheWithBuntdb, t)
}

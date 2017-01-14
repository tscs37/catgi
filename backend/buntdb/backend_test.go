package buntdb

import (
	"testing"

	"git.timschuster.info/rls.moe/catgi/backend/common"
)

func TestRun(t *testing.T) {
	ctx := common.GetTestCtx()
	inmemDB, err := NewBuntDBBackend(map[string]interface{}{
		"file": ":memory:",
	}, ctx)

	if err != nil {
		t.Log("Error on creating Testing Backend: ", err)
		t.FailNow()
		return
	}

	common.RunTestSuite(inmemDB, t)
}

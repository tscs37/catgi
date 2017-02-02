package buntdb

import (
	"testing"

	"git.timschuster.info/rls.moe/catgi/backend/compl_test"
)

func TestCompliance(t *testing.T) {
	ctx := compltest.GetTestCtx()
	inmemDB, err := NewBuntDBBackend(map[string]interface{}{
		"file": ":memory:",
	}, ctx)

	if err != nil {
		t.Log("Error on creating Testing Backend: ", err)
		t.FailNow()
		return
	}

	compltest.RunTestSuite(inmemDB, t)
}

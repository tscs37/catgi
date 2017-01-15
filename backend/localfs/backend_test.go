package localfs

import (
	"fmt"
	"testing"

	"time"

	"os"

	"git.timschuster.info/rls.moe/catgi/backend/compl_test"
)

func TestCompliance(t *testing.T) {
	ctx := compltest.GetTestCtx()
	tmpDir := "/tmp/test-catgi-" + fmt.Sprintf("%d", time.Now().Unix()) + "/"
	os.MkdirAll(tmpDir, 0700)
	t.Log("Using TmpDir: ", tmpDir)
	localfs, err := NewLocalFSBackend(map[string]interface{}{
		"root":     tmpDir,
		"abs_root": true,
	}, ctx)

	if err != nil {
		t.Log("Error on creating Testing Backend: ", err)
		t.FailNow()
		return
	}

	t.Skip("Skipping test due to incomplete implementation.")
	compltest.RunTestSuite(localfs, t)
}

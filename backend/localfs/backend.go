package localfs

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/backend/common"
)

func (l *LocalFSBackend) Name() string { return "localfs" }

func (l *LocalFSBackend) Upload(name string, file *common.File, ctx context.Context) error {
	return common.ErrorNotImplemented
}

func (l *LocalFSBackend) Exists(name string, ctx context.Context) error {
	return common.ErrorNotImplemented
}

func (l *LocalFSBackend) Get(name string, ctx context.Context) (*common.File, error) {
	return nil, common.ErrorNotImplemented
}

func (l *LocalFSBackend) Delete(name string, ctx context.Context) error {
	return common.ErrorNotImplemented
}

func (l *LocalFSBackend) ListGlob(ctx context.Context, glob string) ([]*common.File, error) {
	return nil, common.ErrorNotImplemented
}

func (l *LocalFSBackend) RunGC(ctx context.Context) ([]common.File, error) {
	return nil, common.ErrorNotImplemented
}

// pingFS checks if the root exists and is writable.
func (l *LocalFSBackend) pingFS() error {
	return common.ErrorNotImplemented
}

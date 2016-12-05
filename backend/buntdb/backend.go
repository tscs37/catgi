package buntdb

import (
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"github.com/tidwall/buntdb"
)

type BuntDBBackend struct {
	db *buntdb.DB
}

func (b *BuntDBBackend) Name() string { return "buntdb-backend" }

func (b *BuntDBBackend) Upload(name string, file *types.File) error { return nil }

func (b *BuntDBBackend) Exists(name string) error { return nil }

func (B *BuntDBBackend) Get(name string) (*types.File, error) { return nil, nil }

func (b *BuntDBBackend) LoadIndex(types.Index) error { return nil }

func (b *BuntDBBackend) StoreIndex(types.Index) error { return nil }

func (b *BuntDBBackend) Delete(name string) error { return nil }

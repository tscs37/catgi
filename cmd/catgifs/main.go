package main

import (
	"context"
	"fmt"

	"bazil.org/fuse"
	"git.timschuster.info/rls.moe/catgi/backend"
	"github.com/tscs37/fuse/fs"
)

func main() {
}

type CatgiFS struct {
	index backend.KVBackend
	data  backend.ContentBackend
}

func (c *CatgiFS) Ctx() context.Context {
	return context.Background()
}

func (c *CatgiFS) Root() (fs.Node, error) {
	return &CatgiFSNode{
		fs: c,
	}, nil
}

type CatgiFSNode struct {
	fs    *CatgiFS
	inode uint64
}

func (c *CatgiFSNode) Attr(ctx context.Context, attr *fuse.Attr) error {
	attrs, err := c.fs.index.Get(ctx, fmt.Sprintf("attr:/%d", c.inode))
	if v, ok := attrs.(fuse.Attr); ok {
		attr.Size = v.Size
		attr.Atime = v.Atime
		attr.Crtime = v.Crtime
		attr.Ctime = v.Ctime
		attr.Flags = v.Flags
	}
	return err
}

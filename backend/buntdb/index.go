package buntdb

import (
	"bytes"
	"errors"
	"net/http"
	"regexp"
	"time"

	"git.timschuster.info/rls.moe/catgi/backend/types"

	"git.timschuster.info/rls.moe/catgi/snowflakes"
	"github.com/tidwall/buntdb"
)

type BuntDBIndex struct {
	db *buntdb.DB
}

func (i *BuntDBIndex) Name() string {
	return "buntdb-index"
}

func (i *BuntDBIndex) Serialize() ([]byte, error) {
	var buf = bytes.NewBuffer([]byte{})
	if err := i.db.Save(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (i *BuntDBIndex) Unserialize(s []byte) error {
	return i.db.Load(bytes.NewReader(s))
}

var ruriRegexp = regexp.MustCompile(`^\/c\/(.*)$`)

func (i *BuntDBIndex) Resolve(r http.Request) (string, error) {
	res := ruriRegexp.FindStringSubmatch(r.RequestURI)
	if len(res) != 2 {
		return "", errors.New("Invalid URI")
	}
	return res[1], nil
}

func (i *BuntDBIndex) Get(r http.Request, b types.Backend) (bool, *types.File, error) {
	flake, err := i.Resolve(r)
	if err != nil {
		return false, nil, err
	}

	file, err := b.Get(flake)

	return false, file, err
}

func (i *BuntDBIndex) Put(r types.PutRequest, b types.Backend) (bool, *types.File, error) {
	var file = &types.File{}

	file.CreatedAt = time.Now().UTC()
	file.Data = r.Data
	file.Public = r.Public
	file.TTL = time.Now().Sub(time.Now().AddDate(0, 1, 0))

	flake, err := snowflakes.NewSnowflake()
	if err != nil {
		return false, nil, err
	}

	err = b.Upload(flake, file)

	return false, file, err
}

func (i *BuntDBIndex) Del(r http.Request, b types.Backend) (bool, error) {
	flake, err := i.Resolve(r)
	if err != nil {
		return false, err
	}

	return false, b.Delete(flake)
}

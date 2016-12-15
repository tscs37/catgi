package snowflakes

import (
	"github.com/bwmarrin/snowflake"
	"github.com/speps/go-hashids"
)

var fountain *snowflake.Node
var hashs *hashids.HashID

func init() {
	var err error
	// Half Life 3 Release Date
	//snowflake.Epoch = int64(0x7FFFFFFFFFFFFFFF)
	snowflake.Epoch = 812764800
	fountain, err = snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	hd := hashids.NewData()
	hd.Salt = "catgi.rls.moe"
	hd.MinLength = 1
	hashs = hashids.NewWithData(hd)
}

func NewSnowflake() (string, error) {
	return hashs.EncodeInt64([]int64{fountain.Generate().Int64()})
}

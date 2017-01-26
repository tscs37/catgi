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

// NewSnowflake returns a new unique ID based on a timestamp
// or an error if there is a problem during encoding
//
// Note that these Snowflakes are not cryptographically safe,
// they are not encrypted, just obfuscated.
func NewSnowflake() (string, error) {
	return hashs.EncodeInt64([]int64{NewRawFlake().Int64()})
}

// NewRawFlake returns a snowflake.ID type, useful when a encoded
// string-flake is not sufficient and, for example, a uint64 or
// []byte is needed.
func NewRawFlake() snowflake.ID {
	return fountain.Generate()
}

package crypto

import (
	"fmt"
	"testing"

	"golang.org/x/crypto/blake2b"

	"github.com/stretchr/testify/assert"
)

var (
	testKeySecret = []byte("This is a test Key")
	testKeyValue  = "B4C971C6DF22E906162C6F313485C7EE" +
		"EFFCCA1E4BA4094D9CD0799E6211B9C4" +
		"4C39D76F3EF8AC20869F71170813B84D" +
		"425AA21A290A89D05D655A3576EBAEB1"
	testSecretData = "01BC6DE8C59E8CAA454526822799995D" +
		"B53A2BF2F61AF3EFEAFAE0DF236B9EA5" +
		"25B5B940675C6791B27EDB3979ADFDBC" +
		"6A050A72B3C792AC7D74DE19DC3BA5EE" +
		"A5E21E1D353974"
)

func TestDeriveKey(t *testing.T) {
	assert := assert.New(t)

	secKey := NewSecretKey(testKeySecret)

	assert.EqualValues(secKey, secKey.MustDeriveKey(), "Empty path must equal secret key")
	assert.EqualValues(fmt.Sprintf("%X", secKey), testKeyValue, "Key must be correctly hashed")

	var nextKey SecretKey
	{
		generator, err := blake2b.New512(secKey[:])
		assert.NoError(err)

		generator.Write([]byte("test"))
		copy(nextKey[:], generator.Sum([]byte{}))
	}

	assert.EqualValues(nextKey, secKey.MustDeriveKey("test"), "One path must match")

	var nextNextKey SecretKey
	{
		generator, err := blake2b.New512(nextKey[:])
		assert.NoError(err)

		generator.Write([]byte("test"))
		copy(nextNextKey[:], generator.Sum([]byte{}))
	}

	assert.EqualValues(nextNextKey, secKey.MustDeriveKey("test", "test"), "Two paths must match")
}

func TestGetNBytes(t *testing.T) {
	assert := assert.New(t)

	secKey := NewSecretKey(testKeySecret)

	secretData := secKey.MustGetNSecretBytes(7100, nil)

	assert.Len(secretData, 7100, "Must have 7100 bytes")

	assert.Contains(fmt.Sprintf("%X", secretData), testSecretData, "Derived Secret must be deterministic")
}

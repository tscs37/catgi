package crypto

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptWhole(t *testing.T) {
	assert := assert.New(t)

	plaintext := "THIS IS A PLAINTEXT!"
	keysecret := "This is my secret key password"

	key := NewSecretKey([]byte(keysecret))

	ciphertext, err := EncryptBytes(key.MustDeriveKey("test1"), []byte(plaintext))

	assert.NoError(err)
	assert.NotEmpty(ciphertext)
	assert.True(len(ciphertext) > len(plaintext))

	recovered, err := DecryptBytes(key.MustDeriveKey("test1"), []byte(ciphertext))

	assert.NoError(err)
	assert.EqualValues(plaintext, recovered)
}

func TestEncryptStream(t *testing.T) {
	assert := assert.New(t)

	key := NewSecretKey([]byte("This is my other secret key password"))

	var plaintext []byte
	{
		rndBuf := make([]byte, 32*1024+609)
		_, err := rand.Read(rndBuf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		plaintext = rndBuf
	}

	cipherstream, errChan, err := EncryptStreamData(key, bytes.NewBuffer(plaintext))

	assert.NoError(err)

	ciphertext, err := ioutil.ReadAll(cipherstream)

	assert.NoError(<-errChan)
	assert.NoError(err)

	assert.NotEmpty(ciphertext)
	assert.True(len(ciphertext) > len(plaintext))

	newPlaintextStream, errChan, err := DecryptStreamData(key, bytes.NewBuffer(ciphertext))

	assert.NoError(err)

	newPlaintext, err := ioutil.ReadAll(newPlaintextStream)

	assert.NoError(<-errChan)
	assert.NoError(err)

	assert.NotEmpty(newPlaintext)

	assert.EqualValues(plaintext, newPlaintext)
}

func TestEncryptBlock(t *testing.T) {
	assert := assert.New(t)

	key := NewSecretKey([]byte("This is my other secret key password"))

	var plaintext []byte
	{
		rndBuf := make([]byte, 32*1024+609)
		_, err := rand.Read(rndBuf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		plaintext = rndBuf
	}

	cipherstream, errChan, err := EncryptBlockData(key, bytes.NewBuffer(plaintext))

	assert.NoError(err)

	ciphertext, err := ioutil.ReadAll(cipherstream)

	assert.NoError(<-errChan)
	assert.NoError(err)

	assert.NotEmpty(ciphertext)
	assert.True(len(ciphertext) > len(plaintext))

	newPlaintextStream, errChan, err := DecryptBlockData(key, bytes.NewBuffer(ciphertext))

	assert.NoError(err)

	newPlaintext, err := ioutil.ReadAll(newPlaintextStream)

	assert.NoError(<-errChan)
	assert.NoError(err)

	assert.NotEmpty(newPlaintext)

	assert.EqualValues(plaintext, newPlaintext)
}

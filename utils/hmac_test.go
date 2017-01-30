package utils

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMACGen(t *testing.T) {
	assert := assert.New(t)

	assert.Empty(nil)

	key := []byte("STR TESTING HMAC KEY")

	data1 := []byte("Testing Data 1")
	data2 := []byte("Tsting Datae 2")

	hmac, err := HMAC(key, bytes.NewReader(data1))

	assert.NoError(err, "Must not return error")
	assert.NotNil(hmac, "HMAC cannot be nil")

	err = VerifyHMAC(hmac, key, bytes.NewReader(data1))
	assert.NoError(err, "Must not return error on equal data")

	err = VerifyHMAC(hmac, key, bytes.NewReader(data2))
	assert.Error(err, "Must return error on mismatching data")

	hmac = append(hmac, hmac...)
	err = VerifyHMAC(hmac, key, bytes.NewReader(data1))
	assert.Error(err, "Must return error when mismatching hmac length")

	hmac = hmac[:len(hmac)/2-1]
	err = VerifyHMAC(hmac, key, bytes.NewReader(data1))
	assert.Error(err, "Must return error when mismatching hmac length")

	hmac[2] = 0x22
	hmac = append(hmac, 0)
	err = VerifyHMAC(hmac, key, bytes.NewReader(data1))
	assert.Error(err, "Must return error on mismatching hmac")
}

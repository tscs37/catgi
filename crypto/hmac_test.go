package crypto

import (
	"bytes"
	"crypto/rand"
	"os"
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

func runTHMACBench(b *testing.B, blocksize int) {
	data := bytes.NewReader(make([]byte, blocksize))
	key := make([]byte, 64)
	_, err := rand.Read(key)
	if err != nil {
		b.Log(err)
		b.Fail()
		return
	}
	b.ReportAllocs()
	b.SetBytes(int64(blocksize))
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = HMAC(key, data)
		data.Seek(0, os.SEEK_SET)
	}
}

func runTVerifyBench(b *testing.B, blocksize int) {
	data := bytes.NewReader(make([]byte, blocksize))
	key := make([]byte, 64)
	_, err := rand.Read(key)
	if err != nil {
		b.Log(err)
		b.Fail()
		return
	}
	hmac, err := HMAC(key, data)
	b.ReportAllocs()
	b.SetBytes(int64(blocksize))
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = VerifyHMAC(hmac, key, data)
		data.Seek(0, os.SEEK_SET)
	}
}

func runTVerifyWrongBench(b *testing.B, blocksize int) {
	data := bytes.NewReader(make([]byte, blocksize))
	key := make([]byte, 64)
	hmac := make([]byte, 64)
	_, err := rand.Read(key)
	if err != nil {
		b.Log(err)
		b.Fail()
		return
	}
	_, err = rand.Read(hmac)
	if err != nil {
		b.Log(err)
		b.Fail()
		return
	}
	b.ReportAllocs()
	b.SetBytes(int64(blocksize))
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = VerifyHMAC(hmac, key, data)
		data.Seek(0, os.SEEK_SET)
	}
}

func BenchmarkKeySizeHMACData(b *testing.B) {
	runTHMACBench(b, 64)
}

func BenchmarkKeySizeVerify(b *testing.B) {
	runTVerifyBench(b, 64)
}

func BenchmarkKeySizeVerifyWrong(b *testing.B) {
	runTVerifyWrongBench(b, 64)
}

func Benchmark4kHMACData(b *testing.B) {
	runTHMACBench(b, 4*1024)
}

func Benchmark4kVerify(b *testing.B) {
	runTVerifyBench(b, 4*1024)
}

func Benchmark4kVerifyWrong(b *testing.B) {
	runTVerifyWrongBench(b, 4*1024)
}

func Benchmark64MHMACData(b *testing.B) {
	runTHMACBench(b, 64*1024*1024)
}

func Benchmark64MVerify(b *testing.B) {
	runTVerifyBench(b, 64*1024*1024)
}

func Benchmark64MVerifyWrong(b *testing.B) {
	runTVerifyWrongBench(b, 64*1024*1024)
}

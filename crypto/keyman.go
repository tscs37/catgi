package crypto

import (
	"bytes"
	"hash"

	"fmt"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/hkdf"
)

type SecretKey [64]byte

func NewSecretKey(master []byte) SecretKey {
	if len(master) < 64 {
		newMaster := blake2b.Sum512(master)
		master = newMaster[:]
	}

	internalKey := [64]byte{}

	copy(internalKey[:], master)

	return SecretKey(internalKey)
}

// DeriveKey will compute a new secret based on the given path
// and the defined secret key.
//
// The path is computed recursively on the result of the previous
// path such that the path computation can be split at any point.
// If no path is specified, then the given master is returned.
//
// Illustration:
// DeriveKey(path1, path2) == DeriveKey(DeriveKey(path1), path2)
func (s SecretKey) DeriveKey(path ...string) (SecretKey, error) {
	if len(path) == 0 {
		return s, nil
	}

	var newSecret SecretKey

	newSecretData, err := HMAC(s[:], bytes.NewBufferString(path[0]))
	if err != nil {
		return newSecret, err
	}

	copy(newSecret[:], newSecretData)

	if len(path) > 1 {
		return newSecret.DeriveKey(path[1:]...)
	}
	return newSecret, nil
}

// MustDeriveKey works like DeriveKey but panics on an error.
func (s SecretKey) MustDeriveKey(path ...string) SecretKey {
	secKey, err := s.DeriveKey(path...)
	if err != nil {
		panic(err)
	}
	return secKey
}

// GetNSecretBytes reads N bytes from a HKDF with an optional salt.
// (Set salt to nil if no salt is used)
func (s SecretKey) GetNSecretBytes(n int, salt []byte) ([]byte, error) {
	getHasher := func() hash.Hash {
		hasher, err := blake2b.New512(nil)
		if err != nil {
			panic(err)
		}
		return hasher
	}
	reader := hkdf.New(getHasher, s[:], salt, nil)

	returnData := make([]byte, n)

	read, err := reader.Read(returnData)
	if err != nil {
		return returnData, err
	}
	if read < n {
		return returnData, fmt.Errorf("Wanted %d but got %d bytes", n, read)
	}
	return returnData, err
}

func (s SecretKey) MustGetNSecretBytes(n int, salt []byte) []byte {
	secBytes, err := s.GetNSecretBytes(n, salt)
	if err != nil {
		panic(err)
	}
	return secBytes
}

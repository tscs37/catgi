package crypto

import (
	"bytes"
	"hash"

	"fmt"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/hkdf"
)

// SecretKey is a special 64 byte long secret that can be used
// to derive an arbitrary amount of subkeys.
// All subkeys are deterministically based on their path and
// their parent, allowing an application or library to reuse
// a secret for many operations in different locations.
// As a bonus, if a subkey is compromised, it's parents are not,
// so you can expose as many keys as you like to less trustworthy
// clients without fearing that someone could forge more keys.
//
// Keep in mind that keys are deterministic, so if a key is compromised
// the specific path used will remain forever compromised for the same
// parent key. You need to change the path or the master secret to
// regain security.
//
// Since this key uses HMACs, you can also utilize it to generate
// encryption keys. If a key and it's path are compromised, an attacker
// cannot forge other keys unless they have knowledge of the parent.
//
// To protect the master, it is recommended to use a path depth of atleast
// 2. Ex.: /myapp/example/org/ would be an acceptable derivative secret
// of a master key.
type SecretKey [64]byte

// NewSecretKey will take the given master secret and generate
// a SecretKey from it.
// If the master is shorter than 64 bytes, it will first be hashed
// using Blake2b.
// If the master is longer than 64 bytes, the excessive data is discarded.
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
// Applications should derive to a depth of 3 path segments
// before externalizing a key, ie transmitting it over network.
//
// The path is computed recursively on the result of the previous
// path such that the path computation can be split at any point.
// If no path is specified, then the given master is returned.
//
// Illustration:
//
// DeriveKey(path1, path2) == DeriveKey(DeriveKey(path1), path2)
//
// SecretKey.DeriveKey() == SecretKey
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
// Keep in mind that HKDF has an upper limit on entropy, you should
// not read excessive amounts of random data from here.
// I recommend not using more than 4096 bytes of data, this has worked
// for my tests so far with good reliability and should be sufficient
// for 99% of applications.
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

// MustGetNSecretBytes works like GetNSecretBytes but will panic
// if it encounters an error.
func (s SecretKey) MustGetNSecretBytes(n int, salt []byte) []byte {
	secBytes, err := s.GetNSecretBytes(n, salt)
	if err != nil {
		panic(err)
	}
	return secBytes
}

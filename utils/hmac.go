package utils

import (
	"errors"
	"io"

	"golang.org/x/crypto/blake2b"
)

// HMAC will read the given reader and return the MAC for the
// given key
func HMAC(key []byte, data io.Reader) ([]byte, error) {
	// We must hash long keys but short keys are padded
	// by blake2b.
	if len(key) > 64 {
		newKey := blake2b.Sum512(key)
		copy(key, newKey[:])
	}
	hasher, err := blake2b.New512(key)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(hasher, data)
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

// VerifyHMAC will take an HMAC and a to-be-verified
// data stream plus key. It will generate the HMAC
// for that datastream and then perform a constant-time
// comparison of the two HMACs.
func VerifyHMAC(hmac, key []byte, data io.Reader) error {
	verifyHMAC, err := HMAC(key, data)
	if err != nil {
		return err
	}
	if len(hmac) != len(verifyHMAC) {
		return errors.New("HMAC differing in length")
	}
	var diffs = 0
	for k := range hmac {
		if hmac[k] != verifyHMAC[k] {
			diffs++
		}
	}
	if diffs > 0 {
		return errors.New("HMAC failed")
	}
	return nil
}

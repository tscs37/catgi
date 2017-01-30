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
func VerifyHMAC(hmac, key []byte, data io.Reader) (err error) {
	var verifyHMAC = make([]byte, len(hmac))
	verifyHMAC, err = HMAC(key, data)

	lenhmac := len(hmac)
	lenvmac := len(verifyHMAC)
	if lenhmac != lenvmac {
		// if macs are differing in length, verify macs
		// against itself to avoid timing attacks.
		if lenvmac > lenhmac {
			verifyHMAC = verifyHMAC[:lenhmac]
		} else {
			verifyHMAC = append(verifyHMAC,
				make([]byte, lenhmac-lenvmac)...)
		}
	}

	// do a constant-time compare
	var result byte
	for k := range hmac {
		result |= hmac[k] ^ verifyHMAC[k]
	}
	// if any differences are found, return error
	if result != 0 || lenhmac != lenvmac || err != nil {
		return errors.New("HMAC failed")
	}

	return nil
}

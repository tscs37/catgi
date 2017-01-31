package utils

import (
	"errors"
	"io"

	"golang.org/x/crypto/blake2b"
)

// HMAC will read the given reader and return the MAC for the
// given key
// As lng as the Blake2b Algorithm can be initialized, this
// function returns a hash.
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
		return hasher.Sum(nil), err
	}
	return hasher.Sum(nil), nil
}

// HMACVerificationFail is returned by VerifyHMAC when the hmac
// given did not match the given datastream.
var HMACVerificationFail = errors.New("HMAC Verification failed")

// VerifyHMAC will take an HMAC and a to-be-verified
// data stream plus key. It will generate the HMAC
// for that datastream and then perform a constant-time
// comparison of the two HMACs.
func VerifyHMAC(hmac, key []byte, data io.Reader) (err error) {
	// Note on functionality
	// This function is never allowed to return early, rather
	// it must perform the mac comparison even if a mac could
	// not be computed.

	var verifyHMAC = make([]byte, len(hmac))
	verifyHMAC, err = HMAC(key, data)

	lenhmac := len(hmac)
	lenvmac := len(verifyHMAC)
	if lenhmac != lenvmac {
		// if the length of the macs mismatch we pad or trim
		// the incoming mac accordingly and continue with the
		// compare. this prevents leaking information to the
		// outside.
		if lenhmac > lenvmac {
			hmac = hmac[:lenvmac]
		} else {
			hmac = append(hmac,
				make([]byte, lenvmac-lenhmac)...)
		}
	}

	// do a constant-time compare
	var result byte
	for k := range hmac {
		result |= hmac[k] ^ verifyHMAC[k]
	}
	// if any differences are found, return error
	var retErr error
	if result != 0 {
		retErr = HMACVerificationFail
	}
	if lenhmac != lenvmac {
		retErr = HMACVerificationFail
	}
	if err != nil {
		retErr = HMACVerificationFail
	}

	return retErr
}

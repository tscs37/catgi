package crypto

import "fmt"

// GetKey is a wrapper for GetKeyFromSalt that generates
// a random salt to be used for the key using a random nonce.
// Internally it combines a PreKey with a randomly generated nonce.
func GetKey(key SecretKey, cipher Cipher, mode CryptMode) ([]byte, []byte, error) {
	salt, err := getNonce(24)
	if err != nil {
		return nil, nil, err
	}
	preKey, err := GetPreKey(key, cipher, mode)
	if err != nil {
		return nil, nil, err
	}
	finalKey, err := GetKeyFromPrekey(preKey, salt)
	return finalKey, salt, err
}

// GetKeyFromPreKey applies the given salt to a PreKey.
// You should not use this to derive from anything but the data
// you get from GetPreKey.
func GetKeyFromPrekey(key SecretKey, salt []byte) ([]byte, error) {
	streamKey, err := key.DeriveKey(fmt.Sprintf("%X", salt))
	if err != nil {
		return nil, err
	}
	cipherKey, err := streamKey.GetNSecretBytes(32, nil)
	if err != nil {
		return nil, err
	}
	return cipherKey, nil
}

// GetPreKey derives a so called prekey. This prekey is used when decrypting
// data streams that use the same cipher and cryptmode, allowing the AEAD
// mode to operate at least-depth while still being able to decrypt as much
// data as possible.
func GetPreKey(key SecretKey, cipher Cipher, mode CryptMode) (SecretKey, error) {
	return key.DeriveKey(
		string(mode), fmt.Sprintf("%X", cipher),
	)
}

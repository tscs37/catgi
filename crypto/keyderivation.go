package crypto

import "fmt"

// GetKey is a wrapper for GetKeyFromSalt that generates
// a random salt to be used for the key using a random nonce.
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

// GetKeyFromSalt will deterministically return a cipher key
// from a given secret key, salt, cipher and mode.
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

func GetPreKey(key SecretKey, cipher Cipher, mode CryptMode) (SecretKey, error) {
	return key.DeriveKey(
		string(mode), fmt.Sprintf("%X", cipher),
	)
}

package crypto

// Notes on how to add new crypto:
//
// If a new cipher becomes default, the value of CipherDefault
// must be changed. Old Ciphers will not be removed as they may
// beused to decrypt old ciphers.
//
// All Encryption automtically uses CipherDefault

// Cipher represents a encryption mode used by this package
// The default cipher is CipherPolyCha20
type Cipher uint16

const (
	CipherDefault = CipherPolyCha20

	// Encrypt and Decrypt using Chacha20 with Poly1305
	CipherPolyCha20 = iota
)

type CryptMode string

const (
	CryptModeWhole  = "fullkey"
	CryptModeStream = "streamkey"
	CryptModeBlock  = "blockkey"
)

const (
	BlockSize       = 4 * 1024
	BlockHeaderSize = 8
)

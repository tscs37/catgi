package crypto

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptStreamData will accept a secret key and a incoming
// io.Reader to read from.
// It will encrypt all incoming data using 4KiB blocks.
// If an error occurs during initialization, it will return a nil
// error channel and nil io.Reader.
// If the initialization is correct, it will start the cipher worker
// in a go routine. The Worker will return an error over the given
// error channel and close the io.Reader on it's side.
// The worker will block waiting for either input or output, it's output
// is not buffered.
// The Encryption will derive all secrets from the given SecretKey
// in such a way that the function DecryptStreamData can decrypt it
// given it has the same key.
func EncryptStreamData(key SecretKey, data io.Reader) (io.Reader, <-chan error, error) {
	streamKey, salt, err := GetKey(key, CipherDefault, CryptModeStream)
	if err != nil {
		return nil, nil, err
	}
	aead, err := chacha20poly1305.New(streamKey)
	if err != nil {
		return nil, nil, err
	}
	pipeReader, pipeWriter := io.Pipe()

	worker := &EncryptWorker{
		AEAD:   aead,
		Salt:   salt,
		Input:  data,
		Output: pipeWriter,
	}
	errorChannel := make(chan error, 1)

	go func(worker *EncryptWorker, writer io.WriteCloser) {
		errorChannel <- worker.CipherStream()
		close(errorChannel)
		writer.Close()
	}(worker, pipeWriter)

	return pipeReader, errorChannel, nil
}

func DecryptStreamData(key SecretKey, data io.Reader) (io.Reader, <-chan error, error) {
	streamKey, err := GetPreKey(key, CipherDefault, CryptModeStream)
	if err != nil {
		return nil, nil, err
	}
	pipeReader, pipeWriter := io.Pipe()

	worker := &DecryptWorker{
		Key:    streamKey,
		Input:  data,
		Output: pipeWriter,
	}

	errorChannel := make(chan error, 1)

	go func(worker *DecryptWorker, writer io.WriteCloser) {
		errorChannel <- worker.DecipherStreamOrBlock()
		close(errorChannel)
		writer.Close()
	}(worker, pipeWriter)

	return pipeReader, errorChannel, nil
}

func EncryptBlockData(key SecretKey, data io.Reader) (io.Reader, <-chan error, error) {
	streamKey, salt, err := GetKey(key, CipherDefault, CryptModeBlock)
	if err != nil {
		return nil, nil, err
	}
	aead, err := chacha20poly1305.New(streamKey)
	if err != nil {
		return nil, nil, err
	}
	pipeReader, pipeWriter := io.Pipe()

	worker := &EncryptWorker{
		AEAD:   aead,
		Salt:   salt,
		Input:  data,
		Output: pipeWriter,
	}
	errorChannel := make(chan error, 1)

	go func(worker *EncryptWorker, writer io.WriteCloser) {
		errorChannel <- worker.CipherBlock()
		close(errorChannel)
		writer.Close()
	}(worker, pipeWriter)

	return pipeReader, errorChannel, nil
}

func DecryptBlockData(key SecretKey, data io.Reader) (io.Reader, <-chan error, error) {
	streamKey, err := GetPreKey(key, CipherDefault, CryptModeBlock)
	if err != nil {
		return nil, nil, err
	}
	pipeReader, pipeWriter := io.Pipe()

	worker := &DecryptWorker{
		Key:    streamKey,
		Input:  data,
		Output: pipeWriter,
	}

	errorChannel := make(chan error, 1)

	go func(worker *DecryptWorker, writer io.WriteCloser) {
		errorChannel <- worker.DecipherStreamOrBlock()
		close(errorChannel)
		writer.Close()
	}(worker, pipeWriter)

	return pipeReader, errorChannel, nil
}

func EncryptBytes(key SecretKey, data []byte) ([]byte, error) {
	wholeKey, salt, err := GetKey(key, CipherDefault, CryptModeWhole)
	if err != nil {
		return nil, err
	}
	aead, err := chacha20poly1305.New(wholeKey)
	if err != nil {
		return nil, err
	}
	var out = bytes.NewBuffer([]byte{})
	worker := &EncryptWorker{
		AEAD:   aead,
		Salt:   salt,
		Input:  bytes.NewReader(data),
		Output: out,
	}
	err = worker.CipherAll()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func DecryptBytes(key SecretKey, data []byte) ([]byte, error) {
	wholeKey, err := GetPreKey(key, CipherDefault, CryptModeWhole)
	if err != nil {
		return nil, err
	}
	var out = bytes.NewBuffer([]byte{})
	worker := &DecryptWorker{
		Key:    wholeKey,
		Input:  bytes.NewReader(data),
		Output: out,
	}
	err = worker.DecipherAll()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func getNonce(size int) ([]byte, error) {
	var nonce = make([]byte, size)
	n, err := rand.Read(nonce)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, errors.New("Could not make nonce")
	}
	return nonce, nil
}

// EncryptedBlock contains an encrypted block of data
// with necessary information to decrypt it again.
// When used in Block-Mode, the maximum size that should be used is 2^16-1
// In Stream- or Whole-Mode the maximum safe size is a bit under 2^32, depending
// on nonce, salt and encoding overhead.
type EncryptedBlock struct {
	Cipher  uint16 `msgpack:"cipher"`  // Cipher is a constant determining the cipher algorithm
	Padding uint16 `msgpack:"padding"` // Padding is the number of bytes added to the data. Whe not used set to 0x0000
	Nonce   []byte `msgpack:"nonce"`   // Nonce is the encryption nonce used
	Salt    []byte `msgpack:"salt"`    // Salt is used to derive the key used for decryption
	Data    []byte `msgpack:"data"`    // Data is the actual ciphertext
}

// This is a non-critical error. It indicates that some block had
// the size 0, containing only a header.
var ErrEmptyBlock = errors.New("Empty Block")

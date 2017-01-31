package crypto

import (
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

// EncryptWorker is used to cipher a stream of data.
// It'll read atleast blockSize of bytes and encrypt them
type EncryptWorker struct {
	AEAD   cipher.AEAD // AEAD cipher used
	Salt   []byte      // Salt used for key derivation
	Input  io.Reader   // Data Input Reader
	Output io.Writer   // Cipher Output Writer
}

// CipherStream will read from input and output a ciphered stream
// of data.
func (c *EncryptWorker) CipherStream() error {
	for {
		err := c.cipherBlock(false)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// CipherBlock works like CipherStream but the output blocks have a
// fixed size, making it more suitable for disk/file operations.
func (c *EncryptWorker) CipherBlock() error {
	for {
		err := c.cipherBlock(true)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func (c *EncryptWorker) cipherBlock(usePadding bool) error {
	var (
		data       = make([]byte, BlockSize)
		eofReached = false
	)
	n, err := io.LimitReader(c.Input, BlockSize).Read(data)

	if err != nil && err != io.EOF {
		return err
	} else if err == io.EOF {
		eofReached = true
		if n == 0 {
			return io.EOF
		}
	}

	data = data[:n]

	var padded uint16 = 0
	if n < BlockSize && usePadding {
		if BlockSize-n > 0xFFFF {
			return errors.New("Maximum Padding exceeded")
		}
		padded = uint16(BlockSize - n)
	}

	cipherData, err := c.cipherData(data, padded)
	if err != nil {
		return err
	}

	_, err = c.Output.Write(cipherData)
	if err != nil {
		return err
	}

	if eofReached {
		return io.EOF
	}
	return nil
}

// Cipher all reads the entire plaintext and ciphers it at once
// and encrypts them in a single block
// When all data is encrypted or an error occurs, the output is closed
// and the error returned.
func (c *EncryptWorker) CipherAll() error {
	data, err := ioutil.ReadAll(c.Input)
	if err != nil {
		return err
	}
	cipherData, err := c.cipherData(data, 0)
	if err != nil {
		return err
	}
	_, err = c.Output.Write(cipherData)
	return err
}

// cipherData encrypts the data and pads it if needed.
func (c *EncryptWorker) cipherData(data []byte, padding uint16) ([]byte, error) {
	nonce, err := getNonce(12)
	if err != nil {
		return nil, err
	}
	block := &EncryptedBlock{
		Cipher:  CipherPolyCha20,
		Salt:    c.Salt,
		Nonce:   nonce,
		Padding: padding,
		Data:    []byte{},
	}

	if padding != 0 {
		data = append(data, make([]byte, padding)...)
	}

	if c.AEAD.NonceSize() != len(block.Nonce) {
		return nil, fmt.Errorf("Nonce was %d bytes but AEAD expected %d bytes", len(block.Nonce), c.AEAD.NonceSize())
	}

	block.Data = c.AEAD.Seal(block.Data, block.Nonce, data, nil)

	cipherBlock, err := msgpack.Marshal(block)
	if err != nil {
		return nil, err
	}

	{
		output := make([]byte, len(cipherBlock)+BlockHeaderSize)
		binary.LittleEndian.PutUint64(output[0:BlockHeaderSize], uint64(len(cipherBlock)))
		copy(output[BlockHeaderSize:], cipherBlock)
		cipherBlock = output
	}

	return cipherBlock, nil
}

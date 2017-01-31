package crypto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/crypto/chacha20poly1305"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type DecryptWorker struct {
	Key    SecretKey
	Input  io.Reader
	Output io.Writer
}

func (c *DecryptWorker) DecipherAll() error {
	cipherData, err := ioutil.ReadAll(c.Input)
	if err != nil {
		return err
	}
	data, err := c.decipherData(cipherData)
	if err != nil {
		return err
	}
	_, err = c.Output.Write(data)
	return err
}

func (c *DecryptWorker) DecipherStreamOrBlock() error {
	for {
		err := c.decipherBlock()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func (c *DecryptWorker) decipherBlock() error {
	var (
		blockHeader = make([]byte, BlockHeaderSize)
		eofReached  = false
	)

	n, err := io.LimitReader(c.Input, BlockHeaderSize).Read(blockHeader)

	if err != nil {
		return err
	} else if n < BlockHeaderSize {
		return io.ErrUnexpectedEOF
	} else if err == io.EOF {
		return io.EOF
	}

	blockSize := binary.LittleEndian.Uint64(blockHeader)

	var block = make([]byte, blockSize+BlockHeaderSize)

	copy(block, blockHeader)

	reader := io.LimitReader(c.Input, int64(blockSize))

	dataBlock, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	copy(block[BlockHeaderSize:], dataBlock)

	if err != nil {
		return err
	}

	data, err := c.decipherData(block)
	if err != nil {
		return err
	}
	_, err = c.Output.Write(data)
	if err != nil {
		return err
	}

	if eofReached {
		return io.EOF
	}
	return err
}

// decipherData will read and decipher the given cipherBlock
// Padding options are applied automatically.
func (c *DecryptWorker) decipherData(data []byte) ([]byte, error) {
	if len(data) < BlockHeaderSize {
		return nil, errors.New("Insufficient data for decryption")
	}
	cipherBlockSize := binary.LittleEndian.Uint64(data[0:BlockHeaderSize])

	if uint64(len(data)-BlockHeaderSize) != cipherBlockSize {
		return nil, fmt.Errorf(
			"Incomplete Ciphertext: Have %d, Want %d Bytes",
			uint64(len(data)-BlockHeaderSize),
			cipherBlockSize)
	}

	if cipherBlockSize == 0 {
		return []byte{}, ErrEmptyBlock
	}

	// Cut out the header
	data = data[BlockHeaderSize:]

	var cipherBlock = &EncryptedBlock{}
	err := msgpack.Unmarshal(data, cipherBlock)
	if err != nil {
		return nil, err
	}

	blockKey, err := GetKeyFromPrekey(c.Key, cipherBlock.Salt)
	if err != nil {
		return nil, err
	}
	aead, err := chacha20poly1305.New(blockKey)
	if err != nil {
		return nil, err
	}

	data, err = aead.Open([]byte{}, cipherBlock.Nonce, cipherBlock.Data, nil)

	if cipherBlock.Padding != 0 {
		if len(data) > int(cipherBlock.Padding) {
			data = data[:(uint16(len(data)) - cipherBlock.Padding)]
		} else {
			return nil, errors.New("Too much padding")
		}
	}

	return data, err
}

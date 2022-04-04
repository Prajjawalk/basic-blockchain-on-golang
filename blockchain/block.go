package blockchain

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

//Block keeps block headers
type Block struct {
	Height         uint64
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
}

// SetHash calculates and sets block hash
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	newHeight := []byte(strconv.FormatUint(b.Height, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp, newHeight}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

// NewBlock creates and return Block
func NewBlock(data, prevBlockHash []byte, height uint64) *Block {
	block := &Block{height, time.Now().Unix(), data, prevBlockHash, []byte{}}
	block.SetHash()
	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock() *Block {
	return NewBlock([]byte("Genesis Block"), []byte{}, 0)
}

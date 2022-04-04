package blockchain

import "sync"

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
    Blocks []*Block
    Chainlock sync.RWMutex
    Chainheight uint64
    Latestblock *Block
}

// AddBlock saves provided data as a block in blockchain
func (bc *Blockchain) AddBlock(data []byte) {
    bc.Chainlock.RLock()
    prevBlock := bc.Blocks[len(bc.Blocks)-1]
    newHeight := prevBlock.Height + uint64(1)
    newBlock := NewBlock(data, prevBlock.Hash, newHeight)
    bc.Blocks = append(bc.Blocks, newBlock)
    bc.Chainheight = newHeight
    bc.Latestblock = newBlock
    bc.Chainlock.RUnlock()
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
    return &Blockchain{Blocks: []*Block{NewGenesisBlock()}, Chainheight: 0, Latestblock: NewGenesisBlock()}
}
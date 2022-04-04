package gossipblocks

import (
	"adafel-blockchain/blockchain"
)

type GossipMessage struct {
	Header *Header
	Payload *Payload
}

type Header struct {
	Height uint64
}

type Payload struct {
	Latestblock *blockchain.Block
}

func NewHeader(bc *blockchain.Blockchain) *Header {
	return &Header{Height : bc.Chainheight}
}

func NewPayload(bc *blockchain.Blockchain) *Payload {
	return &Payload{Latestblock : bc.Latestblock}
}

func NewGossipBlock(bc *blockchain.Blockchain) *GossipMessage {
	return &GossipMessage{
		Header : NewHeader(bc),
		Payload : NewPayload(bc),
	}
} 
package main

import (
	"fmt"
	"sync"
	"strings"
	//"log"
	"adafel-blockchain/gossip"
	"adafel-blockchain/gossip/kademlia"
	//"go.uber.org/zap"
)

// ChatMessage is an example struct that is registered on example nodes, and serialized/deserialized on-the-fly.
type ChatMessage struct {
	content string
}

// Marshal serializes a chat message into bytes.
func (m ChatMessage) Marshal() []byte {
	return []byte(m.content)
}

// Unmarshal deserializes a slice of bytes into a chat message, and returns an error should deserialization
// fail, or the slice of bytes be malformed.
func UnmarshalChatMessage(buf []byte) (ChatMessage, error) {
	return ChatMessage{content: strings.ToValidUTF8(string(buf), "")}, nil
}

func main() {

	ch1 := make(chan bool)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	server, err := gossip.NewNode(gossip.WithNodeBindPort(6060))
	if err != nil {
		panic(err)
	}

	defer server.Close()
	
	server.RegisterMessage(ChatMessage{}, UnmarshalChatMessage)

	kserver := kademlia.New()

	server.Bind(kserver.Protocol())

	if err := server.Listen(); err != nil {
		panic(err)
	}
	fmt.Printf("The address of server is %v\n", server.Addr())

    go func() {
		server.Handle(func(ctx gossip.HandlerContext) error {
			obj, err := ctx.DecodeMessage()
			if err != nil {
				return nil
			}
	
			msg, ok := obj.(ChatMessage)
			if !ok {
				return nil
			}
	
			fmt.Printf("Got a message from Bob: '%s'\n", msg.content)
	
			return nil
		})

		ch1 <- true
		wg.Done()
	}()

	// wait until WaitGroup is done
	wg.Wait()
	<- ch1
}
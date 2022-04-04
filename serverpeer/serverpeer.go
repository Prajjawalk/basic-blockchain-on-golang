package main

import (

	"fmt"
	"sync"
	"bytes"
	"log"
	"net"
	//"time"
	"strings"
	"context"
	"encoding/json"
	"encoding/binary"
	//"encoding/gob"
	//"encoding/hex"
	"net/http"
	"io/ioutil"

	"adafel-blockchain/gossip"
	//"adafel-blockchain/gossipblocks"
	"adafel-blockchain/gossip/kademlia"
	//"adafel-blockchain/multicast"
	"adafel-blockchain/blockchain"
	"go.uber.org/zap"
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
/*
const (
	address = "239.0.0.0:9999"
)
*/
var blockHeight uint64

var bc = blockchain.NewBlockchain()

func main() {

	blockHeight = bc.Chainheight
	
	ch1 := make(chan bool)
	ch2 := make(chan bool)
	ch3 := make(chan bool)

	wg := new(sync.WaitGroup)
	wg.Add(3)
	
	peer, err := gossip.NewNode(gossip.WithNodeBindHost(net.ParseIP("0.0.0.0")), gossip.WithNodeLogger(zap.NewExample()))
	fmt.Println("the node is initiated")

	if err != nil {
		panic(err)
	}
	kpeer := kademlia.New()

	defer peer.Close()

	peer.RegisterMessage(ChatMessage{}, UnmarshalChatMessage)

	peer.Bind(kpeer.Protocol())
	
	http.HandleFunc("/", handler)

	http.HandleFunc("/foo", blockhandler)

	go func() {
		http.ListenAndServe("localhost:9000", nil)
		wg.Done()
		ch1 <- true
	}()

	go func() {
		
		if _, err := peer.Ping(context.TODO(), "0.0.0.0:6060"); err != nil {
			log.Panic(err)
		}
		/*
		if err := peer.SendMessage(context.TODO(), "0.0.0.0:6060", ChatMessage{content: "Hi server!"}); err != nil {
			log.Panic(err)
		} 
		
		fmt.Printf("The peers discovered are: %v\n", kpeer.Discover())
		
		peer.Handle(func(ctx gossip.HandlerContext) error {
			fmt.Println("Handling")
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
		*/
		ch2 <- true
		wg.Done()
	}()
		
	go func() {
		/*
		conn, err := multicast.NewBroadcaster(address)
		if err != nil {
			log.Fatal(err)
		}
		
		for {
			gossipblock := gossipblocks.NewGossipBlock(bc)
			var postblock bytes.Buffer

			enc := gob.NewEncoder(&postblock)

			err = enc.Encode(gossipblock)
			if err != nil {
				log.Fatal("encode error:", err)
			}
			conn.Write(postblock.Bytes())
			time.Sleep(1 * time.Second)
		}
		
		peer.Handle(func(ctx gossip.HandlerContext) error {
			if !ctx.IsRequest() {
				return nil
			}
	
			//var r []byte
			
			return ctx.Send([]byte("bye"))
		})
		//fmt.Println("handled")
		*/
		ch3 <- true
		wg.Done()
	}()
	
	if err := peer.Listen(); err != nil {
		panic(err)
	}
	
	//multicast.Listen(address, msgHandler)
	
	peer.Handle(func(ctx gossip.HandlerContext) error {
		if !ctx.IsRequest() {
            return nil
		}

		height, _ := binary.Uvarint(ctx.Data())
		if height < bc.Chainheight {
			blockArray := bc.Blocks[height + 1]

			postblock, err := json.Marshal(blockArray)
			if err != nil {
				log.Fatal(err)
			}

			//postBlockBody := bytes.NewBuffer(postblock)
			return ctx.Send(postblock)
		} else {
			return nil
		}
		
		//return ctx.Send([]byte("bye"))
	})
	
	wg.Wait()
	<- ch1
	<- ch2
	<- ch3
}
/*
func msgHandler(src *net.UDPAddr, b []byte) {
	//fmt.Printf("bytes read from %v", src.String())
	
	var outblock *gossipblocks.GossipMessage

	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)

	err := dec.Decode(&outblock)
    if err != nil {
        log.Fatal(err)
    }
	//fmt.Printf("The received block is %v\n", outblock.Header.Height)
}
*/
func handler(w http.ResponseWriter, r *http.Request) {
	n, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var height uint64

	err = json.Unmarshal(n, &height)
	if err != nil {
		log.Fatal(err)
	}

	if height > blockHeight {

		x, err := json.Marshal(blockHeight)
		if err != nil {
			log.Fatal(err)
		}
		
		_, err = w.Write(x)
		if err != nil {
			panic(err)
		}
	}
}

func blockhandler(w http.ResponseWriter, r *http.Request) {
	n, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var s []*blockchain.Block

	err = json.Unmarshal(n, &s)
	if err != nil {
		log.Fatal(err)
	}

	fileStruct := new(bytes.Buffer)

	err = json.NewEncoder(fileStruct).Encode(s)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(s); i++ {
		bc.Chainlock.RLock()
		bc.Blocks = append(bc.Blocks, s[i])
		bc.Chainheight = s[i].Height
		bc.Latestblock = s[i]
		bc.Chainlock.RUnlock()
	} 

	nonceString := fmt.Sprintf("%v", bc.Chainheight)

	err = ioutil.WriteFile("./block" + nonceString + ".dat", fileStruct.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
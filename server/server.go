package main

import (
	"encoding/json"
	"fmt"
	"bytes"
	"io/ioutil"
	"log"
	//"context"
	"sync"
	//"net"
	//"encoding/binary"
	"net/http"
	//"strconv"
	//"sync"
	"strings"
	"adafel-blockchain/blockchain"
	//"adafel-blockchain/cuckoo"
	//"adafel-blockchain/gossip"
	//"adafel-blockchain/gossip/kademlia"
	"github.com/syndtr/goleveldb/leveldb"
	//"go.uber.org/zap"
)

var bc = blockchain.NewBlockchain()
//var c = cuckoo.NewCuckoo(10, 0.1)
var db, err = leveldb.OpenFile("/mnt/e/basic-blockchain-on-golang/db", nil)

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
	
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := iter.Key()
		value := iter.Value()
		fmt.Printf("{%v, %v}\n", string(key), string(value))
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		panic(err)
	}
	
	ch1 := make(chan bool)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	
    // create a default route handler
	http.HandleFunc( "/", handler)

	// goroutine to launch a server on port 8080
    go func() {
        log.Fatal( http.ListenAndServe( ":8080", nil ) )
		ch1 <- true // one goroutine finished
		wg.Done()
    }()

	

	// wait until WaitGroup is done
	wg.Wait()
	<- ch1

	
}

func handler(w http.ResponseWriter, r *http.Request) {
	// try to read data from the connection

	n, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var s map[string]string
	err = json.Unmarshal(n, &s)
	if err != nil {
		log.Fatal(err)
	}
	
	block_data, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}
	bc.AddBlock(block_data)
	db.Put([]byte(s["address"]), []byte(s["uuid"]), nil)

	// send a response
	var str = []string{"Approved!"}
	var x = []byte{}
	// convert string array to byte array so it can
	// be written to the connection
	for i := 0; i < len(str); i++ {
		b := []byte(str[i])
		for j := 0; j < len(b); j++ {
			x = append(x, b[j])
		}
	}
	// write the data to the connection
	_, err = w.Write(x)
	if err != nil {
		panic(err)
	}

		
	postBody, _ := json.Marshal(bc.Chainheight)

	requestBody := bytes.NewBuffer(postBody)

	fmt.Printf("the request is %v \n", requestBody)

	resp, err := http.Post("http://localhost:9000", "application/json", requestBody)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var height uint64
	
	err = json.Unmarshal(body, &height)
	if err != nil {
		log.Fatal(err)
	}
		
	blockArray := bc.Blocks[height + 1:]

	postblock, err := json.Marshal(blockArray)
	if err != nil {
		log.Fatal(err)
	}

	postBlockBody := bytes.NewBuffer(postblock)
		
	_, err = http.Post("http://localhost:9000/foo", "application/json", postBlockBody)
		
	if err != nil {
		log.Fatal(err)
	}


		

		
	for _, block := range bc.Blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
		
	//}

}


 package main

import (

	"fmt"
	"sync"
	"log"
	"net"
	"strings"
	"context"
	"io/ioutil"
	//"time"
	"net/http"
	"bytes"
	"encoding/gob"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"crypto/x509"
	//"crypto/ecdsa"
	"crypto/sha256"
	//"crypto/rand"
	//"encoding/hex"

	"adafel-blockchain/wallet"
	"adafel-blockchain/gossip"
	"adafel-blockchain/gossipblocks"
	"adafel-blockchain/user"
	"adafel-blockchain/blockchain"
	"adafel-blockchain/gossip/kademlia"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"github.com/ethereum/go-ethereum/crypto"
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

//const (
//	address = "224.0.0.251:9999"
	//address1 = "106.66.224.153:5353"
//)

type server struct {
	user.UnimplementedUserServer
}

type LoginInfo struct {
	ID string
	Key []byte
}

var mywallet *wallet.Wallet

var Email string

// GetDetails implements user.UserServer
func (s *server) GetDetails(ctx context.Context, in *user.UserRequest) (*user.UserResponse, error) {
	Email = in.Email
	fmt.Printf("the email is :%v", Email)
	mywallet = wallet.GetWallet(Email)
	fmt.Printf("The wallet address is %v\n", string(mywallet.GetAddress()))
	return &user.UserResponse{Ok: "Approved!"}, nil
}

func (s *server) Login(ctx context.Context, in *user.WebUrl) (*user.AppResponse, error) {
	Url := in.GetUrl()
	Aid := in.GetAid()
	fmt.Println(Url)
	//content, err := ioutil.ReadFile("priv-key.pem")
	
	privkey, err := crypto.LoadECDSA("priv-key.pem")
	if err != nil {
		log.Fatal(err)
	}
	
	/*
	block, _ := pem.Decode(content)
	if block == nil || block.Type != "ECDSA PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing public key")
	}
	privkey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	*/
	hash := sha256.Sum256([]byte("this is sample data"))
	//signed_priv, err := ecdsa.SignASN1(rand.Reader, privkey.(*ecdsa.PrivateKey), hash[:])
	signed_priv, err := crypto.Sign(hash[:], privkey)
	if err != nil {
		log.Fatal(err)
	}

	data := LoginInfo {
		ID : Url,
		Key : signed_priv,
	}

	input, err := json.Marshal(data)

	_, err = http.Post("http://localhost:10080/adafel/login-with-adafel/"+Aid, "application/json", bytes.NewReader(input))
		
	if err != nil {
		log.Fatal(err)
	}

	return &user.AppResponse{Res: "sent!"}, nil
}

var bc = blockchain.NewBlockchain()

var mystate = gossipblocks.NewGossipBlock(bc)

func main() {
	
	ch1 := make(chan bool)
	ch2 := make(chan bool)
	ch3 := make(chan bool)

	wg := new(sync.WaitGroup)
	wg.Add(3)

	
	go func() {
		lis, err := net.Listen("tcp", "localhost:4500")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		user.RegisterUserServer(grpcServer, &server{})
		grpcServer.Serve(lis)
		ch1 <- true //one goroutine finished
		wg.Done()
	}()


	peer, err := gossip.NewNode(gossip.WithNodeBindHost(net.ParseIP("0.0.0.0")), gossip.WithNodeLogger(zap.NewExample()))
	if err != nil {
		panic(err)
	}
	kpeer := kademlia.New()

	defer peer.Close()

	peer.RegisterMessage(ChatMessage{}, UnmarshalChatMessage)

	peer.Bind(kpeer.Protocol())
	signup()
	
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
		for {
			r := make([]byte, binary.MaxVarintLen64)
			binary.PutUvarint(r, bc.Chainheight)
			for _, dpeer := range(kpeer.Discover()) {
				if dpeer.Port != 6060 {
					//r := make([]byte, binary.MaxVarintLen64)
					//binary.PutUvarint(r, bc.Chainheight)

					response, err := peer.Request(context.TODO(), dpeer.Address, r)
					if err != nil {
						log.Fatal(err)
					}
					if response != nil { 
						var s *blockchain.Block

						err = json.Unmarshal(response, &s)
						if err != nil {
							log.Fatal(err)
						}

						fileStruct := new(bytes.Buffer)

						err = json.NewEncoder(fileStruct).Encode(s)
						if err != nil {
							log.Fatal(err)
						}

						//for i := 0; i < len(s); i++ {
						bc.Chainlock.RLock()
						bc.Blocks = append(bc.Blocks, s)
						bc.Chainheight = bc.Chainheight + uint64(1)
						bc.Latestblock = s
						bc.Chainlock.RUnlock()
						//} 

						err = ioutil.WriteFile("./block" + ".dat", fileStruct.Bytes(), 0755)
						if err != nil {
							log.Fatal(err)
						}
					}
				} 
		
			}
		}
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
			var postblock bytes.Buffer

			enc := gob.NewEncoder(&postblock)
			err = enc.Encode(mystate)
			if err != nil {
				log.Fatal("encode error:", err)
			}
			conn.Write(postblock.Bytes())
			time.Sleep(1 * time.Second)
		}
		*/
		/*
		for _, dpeer := range(kpeer.Discover()) {
			if dpeer.Port != 6060 {
				//var r []byte

				response, err := peer.Request(context.TODO(), dpeer.Address, []byte("hello"))
				if err != nil {
					log.Fatal(err)
				}
				if response != nil { 
					fmt.Printf("the received response is %v\n", response)
				}
			} else {
				fmt.Println("only 6060")
			}
	
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
	
	peer.Handle(func(ctx gossip.HandlerContext) error {
		if !ctx.IsRequest() {
            return nil
		}

		height, _ := binary.Uvarint(ctx.Data()) //response of the peer block height
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
	
	//multicast.Listen(address, msgHandler)
	
	
	wg.Wait()
	<- ch1
	<- ch2
	<- ch3
}

func msgHandler(src *net.UDPAddr, b []byte) {
	//fmt.Printf("bytes read from %v", src.String())

	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)

	err := dec.Decode(&mystate)
    if err != nil {
        log.Fatal(err)
	}
	
	//fmt.Printf("The received block is %v\n", mystate.Header.Height)
}

func signup() {
	content, err := ioutil.ReadFile("pub-key.pub")
	if err != nil {
		log.Fatal(err)
	}
	block, rest := pem.Decode(content)
	if block == nil || block.Type != "ECDSA PUBLIC KEY" {
		log.Fatal("failed to decode PEM block containing public key")
	}
	publickey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	publickey2 := publickey

	fmt.Printf("Got a %T, with remaining data: %q", publickey2, rest)

}


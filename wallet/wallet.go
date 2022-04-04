package wallet

import (
	"bytes"
	"crypto/ecdsa"
	//"crypto/elliptic"
	//"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/ethereum/go-ethereum/crypto"
)

const version = byte(0x00)
const addressChecksumLen = 4

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
	uuid       string
	approved   bool
}

func GetWallet(uuid string) *Wallet {

	private, public := newKeyPair()

	//var uuid string

	//fmt.Scanln(&uuid)

	wallet := Wallet{private, public, uuid, false}

	address := wallet.GetAddress()

	var s string
	fmt.Printf("%v\n", s)

	for s != "Approved!" {
		//fmt.Println("Enter ID")
		//fmt.Scanln(&uuid)

		postBody, _ := json.Marshal(map[string]string{
			"address": string(address),
			"uuid":    string(uuid),
		})

		responseBody := bytes.NewBuffer(postBody)

		resp, err := http.Post("http://localhost:8080", "application/json", responseBody)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()

		//Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		s = string(body)
		fmt.Println(s)
	}
	
	wallet.uuid = uuid
	wallet.approved = true
	fmt.Println("success!")
	return &wallet
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()

	var uuid string

	fmt.Scanln(&uuid)

	wallet := Wallet{private, public, uuid, false}

	return &wallet
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)
	publicSHA2562 := publicSHA256[:]

	return publicSHA2562
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	//curve := elliptic.P256()
	//private, err := ecdsa.GenerateKey(curve, rand.Reader)
	private, err := crypto.GenerateKey()
	if err != nil {
		log.Panic(err)
	}
	
	/*
	privkey_bytes, err := x509.MarshalPKCS8PrivateKey(private)
	if err != nil {
		log.Panic(err)
	}
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
				Type:  "ECDSA PRIVATE KEY",
				Bytes: privkey_bytes,
		},
	)
	ioutil.WriteFile("priv-key.pem", privkey_pem, 0755)
	fmt.Println("saved private key")
	*/
	
	err = crypto.SaveECDSA("priv-key.pem", private)
	if err != nil {
		log.Panic(err)
	} 
	fmt.Println("created private key")
	
	publickey := &private.PublicKey
	fmt.Printf("%T\n", publickey);

	pubkey_bytes, err := x509.MarshalPKIXPublicKey(publickey)
	if err != nil {
		log.Panic(err)
	}
	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
				Type:  "ECDSA PUBLIC KEY",
				Bytes: pubkey_bytes,
		},
	)
	ioutil.WriteFile("pub-key.pub", pubkey_pem, 0755)

	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}

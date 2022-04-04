package main

import (
	"fmt"
	"io"
	"sync"
	"net/http"
	//"net"
	"os"
	//"bytes"
	"log"
	"crypto/x509"
	"crypto/sha256"
	"crypto/ecdsa"
	"io/ioutil"
	"text/template"
	"encoding/pem"
	"encoding/json"

	//"adafel-blockchain/node"

	"github.com/syndtr/goleveldb/leveldb"
)

// Compile templates on start of the application
var templates = template.Must(template.ParseFiles("./public/upload.html", "./public/signin.html", "./public/loggedin.html"))

var c = false
type LoginInfo struct {
	ID string
	Key []byte
}

// Display the named template
func display(w http.ResponseWriter, page string, data interface{}) {
	templates.ExecuteTemplate(w, page+".html", data)
}

var db, err = leveldb.OpenFile("./user-publickeys/db", nil)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		display(w, "upload", nil)
	case "POST":
		uploadFile(w, r)
	}
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//display(w, "signin", nil)
		if c == true {
			c = false
			http.Redirect(w, r, "http://localhost:10080/successfully-logged-in", 303)
		} else {
			display(w, "signin", nil)
		}
	/*
	case "POST": 
		output, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", output)
		http.Redirect(w, r, "http://google.com", 303)
		//display(w, "loggedin", nil)
	*/
	}
}

func loggedinHandler(w http.ResponseWriter, r *http.Request) {
	display(w, "loggedin", nil)
	
}

func main() {
	ch1 := make(chan bool)
	wg := new(sync.WaitGroup)
	wg.Add(1)


	//http.Handle("/assets/", http.StripPrefix("http://localhost:10080/assets/", http.FileServer(http.Dir("E:\\basic-blockchain-on-golang\\client\\public/assets")))) 

	// Signup route
	http.HandleFunc("/signup-with-adafel", uploadHandler)

	//Signin route
	http.HandleFunc("/signin-with-adafel", signinHandler)

	//Adafel Login route
	http.HandleFunc("/login-with-adafel", login)

	//Loggedin route
	http.HandleFunc("/successfully-logged-in", loggedinHandler)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("/mnt/e/basic-blockchain-on-golang/client/public/assets")))) 


	go func() {
		//Listen on port 10080
		http.ListenAndServe(":10080", nil)
		wg.Done()
		ch1 <- true
	}()

	wg.Wait()
	<- ch1
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(10 << 20)
	email := r.FormValue("email")
	fmt.Println(email)
	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create file
	dst, err := os.Create(handler.Filename)
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully Uploaded File\n")
	signup(email)
}

func login(w http.ResponseWriter, r *http.Request) {
	
	output, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var data LoginInfo

	json.Unmarshal(output, &data)
	email := data.ID
	signed_priv := data.Key
	
	//email := r.FormValue("email")
	input := string(email)
	//fmt.Printf("%T",input)
	block, err := db.Get([]byte(input), nil)
	pubkey, err := x509.ParsePKIXPublicKey(block)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%v", pubkey)
	hash := sha256.Sum256([]byte("this is sample data"))
	bool := ecdsa.VerifyASN1(pubkey.(*ecdsa.PublicKey), hash[:], signed_priv)
	//fmt.Printf("%v", bool)
	if bool {
		fmt.Println("true")
		c = true
		//http.Redirect(w, r, "http://google.com", 301)
		//fmt.Fprintf(w, "Successfully Logged in\n")
		//_, err = http.Post("http://localhost:10080/signin-with-adafel", "application/json", bytes.NewReader([]byte("adcsv")))
	}
}

func signup(email string) {
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

	fmt.Printf("Got a %T, with remaining data: %q", publickey, rest)
	db.Put([]byte(email), block.Bytes, nil)

}
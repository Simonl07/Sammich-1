package p3

import (
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"../p3/data"
)

// Chain
var SBC data.SyncBlockChain

// In memory data structures
var identityMap map[int32]data.Identity
var userPubKeyMap map[int32]rsa.PublicKey
var compPubKeyMap map[string]rsa.PublicKey

//Caches for acceptance and application
var applicationCache []data.Submission
var acceptanceCache map[string]int32

// UID count
var UID int32

// init will be executed before everything else.
// Some initialization will be done here.
func init() {
	SBC = data.NewBlockChain()
	identityMap = make(map[int32]data.Identity)
	userPubKeyMap = make(map[int32]rsa.PublicKey)
	compPubKeyMap = make(map[string]rsa.PublicKey)
	// First 0-99 are reserved for potential testing
	UID = 99
}

// Apply submits the application for a given user
func Apply(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	sub, err2 := data.DecodeSubmissionJson(body)
	if err2 != nil {
		w.WriteHeader(500)
		return
	}
	// TODO: Verify Nonce
	uid := generateUID()
	identityMap[uid] = sub.Id
	userPubKeyMap[uid] = sub.PubKey
	cache = append(cache, sub)
	if len(cache) > 3 {

	}
}

func addToChain() {
	SBC.AddToChain(cache)
}

func startAddition() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			addToChain()
		}
	}()
	// ticker.Stop()
}

// Fetch list of job applications
func FetchApplications(w http.ResponseWriter, r *http.Request) {

}

// Register a business and their public key
func RegisterBusiness(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	var reg Registration
	err = json.Unmarshal(body, &reg)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	compPubKeyMap[reg.CompanyName] = reg.PublicKey
}

// Accept a user
func Accept(w http.ResponseWriter, r *http.Request) {

}

// Show Blockchain
func Show(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(SBC.Show()))
	if err != nil {
		w.WriteHeader(500)
	}
	w.WriteHeader(200)
}

// Download Blockchain
func Download(w http.ResponseWriter, r *http.Request) {
	jsonString, err := json.Marshal(&SBC)
	if err != nil {
		w.WriteHeader(500)
	}
	_, err2 := w.Write([]byte(jsonString))
	if err2 != nil {
		w.WriteHeader(500)
	}
	w.WriteHeader(200)
}

func generateUID() int32 {
	UID = UID + 1
	return UID
}

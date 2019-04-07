package p3

import (
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"../p2"
	"../p3/data"
)

// Chain
var SBC data.SyncBlockChain

// In memory data structures
var identityMap map[int32]data.Identity
var userPubKeyMap map[int32]rsa.PublicKey
var compPubKeyMap map[string]rsa.PublicKey

//Caches for acceptance and application
var applicationCache map[int32]Merit
var acceptanceCache map[string]int32
var CACHE_THRESHOLD = 3

// UID count
var UID int32

// init will be executed before everything else.
// Some initialization will be done here.
func init() {

	// init sbc
	SBC = data.NewBlockChain()

	// init data structures
	identityMap = make(map[int32]data.Identity)
	userPubKeyMap = make(map[int32]rsa.PublicKey)
	compPubKeyMap = make(map[string]rsa.PublicKey)

	// init Caches
	applicationCache = make(map[int32]Merit)
	acceptanceCache = make(map[string]int32)
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
	applicationCache[uid] = sub.Merit
	if len(application) > CACHE_THRESHOLD {
		flushCache2BC()
	}
}

func flushCache2BC() {
	acceptMpt := p1.MerklePatriciaTrie{}
	acceptMpt.Initial()
	applyMpt := p1.MerklePatriciaTrie{}
	applyMpt.Initial()

	for k, v := range applicationCache {
		applyMpt.Insert(k, v)
		delete(applicationCache, k)
	}

	for k, v := range acceptanceCache {
		acceptMpt.Insert(k, v)
		delete(acceptanceCache, k)
	}
	block := make(p2.Block)
	if sbc.bc.Length == 0 {
		block.Initial(sbc.bc.Length+1, "GENESIS", acceptMpt, applyMpt)
	} else {
		block.Initial(sbc.bc.Length+1, sbc.Get(sbc.bc.Length - 1)[0].Header.Hash, acceptMpt, applyMpt)
	}

	sbc.Insert(block)
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

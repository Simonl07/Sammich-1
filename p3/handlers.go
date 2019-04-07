package p3

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"../p1"
	"../p2"
	"./data"
	"github.com/gorilla/mux"
)

// Chain
var SBC data.SyncBlockChain

// In memory data structures
var identityMap map[int32]data.Identity
var userPubKeyMap map[int32]string
var compPubKeyMap map[string]string

//Caches for acceptance and application
var applicationCache map[int32]data.Merit
var acceptanceCache map[string]int32
var cachemux sync.Mutex

// UID count
var UID int32

// init will be executed before everything else.
// Some initialization will be done here.
func init() {

	// init sbc
	SBC = data.NewBlockChain()

	// init data structures
	identityMap = make(map[int32]data.Identity)
	userPubKeyMap = make(map[int32]string)
	compPubKeyMap = make(map[string]string)

	// init Caches
	applicationCache = make(map[int32]data.Merit)
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
	// TODO: BEFORE ALL THIS, we still need to verify this is a valid application
	// VERIFY: signature & nonce
	uid := generateUID()
	identityMap[uid] = sub.Id
	userPubKeyMap[uid] = sub.PubKey
	applicationCache[uid] = sub.Merit
}

func flushCache2BC() {
	acceptMpt := p1.MerklePatriciaTrie{}
	acceptMpt.Initial()
	applyMpt := p1.MerklePatriciaTrie{}
	applyMpt.Initial()

	for k, v := range applicationCache {
		inchainMerit := new(data.InchainMerit)
		inchainMerit.Skills = v.Skills
		inchainMerit.Education = v.Education
		inchainMerit.Experience = v.Experience
		inchainMerit.UID = k
		inchainMeritJSON, err := json.Marshal(inchainMerit)
		if err != nil {
			fmt.Print("UNABLE TO FLUSH CACHE TO BC")
		}
		applyMpt.Insert(string(k), string(inchainMeritJSON))
		delete(applicationCache, k)
	}

	for k, v := range acceptanceCache {
		acceptMpt.Insert(k, string(v))
		delete(acceptanceCache, k)
	}
	block := new(p2.Block)
	if SBC.Length() == 0 {
		block.Initial(SBC.Length()+1, "GENESIS", acceptMpt, applyMpt)
	} else {
		parentBlock, _ := SBC.Get(SBC.Length() - 1)
		block.Initial(SBC.Length()+1, parentBlock[0].Header.Hash, acceptMpt, applyMpt)
	}

	SBC.Insert(*block)
}

func startAddition() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			flushCache2BC()
		}
	}()
	// ticker.Stop()
}

// Fetch list of job applications
func FetchMerits(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < int(SBC.Length()); i++ {
		//block := sbc.Get(i)
		//TODO
	}
}

// Register a business and their public key
func RegisterBusiness(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	var reg data.Registration
	err = json.Unmarshal(body, &reg)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	compPubKeyMap[reg.CompanyName] = reg.PubKey
}

// Accept a user
func Accept(w http.ResponseWriter, r *http.Request) {
	/*
		1. Verify(optional):
			m = url path
			make sure that H(m) == decrypt(signature, pub)
		2. Add acceptance to cache
		3. Respond with Identity + PubKey of applicant
	*/
	vars := mux.Vars(r)
	// 1
	// 2
	company := vars["company"]
	uidTemp, err := strconv.Atoi(vars["uid"])
	uid := int32(uidTemp)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	acceptanceCache[company] = uid
	// 3
	identity, oki := identityMap[uid]
	publicKey, okp := userPubKeyMap[uid]
	if !oki || !okp {
		fmt.Print("UNABLE TO GET USER INFO")
	}
	type info struct {
		Idt data.Identity `json:"identity"`
		Pk  string        `json:"publicKey"`
	}
	jsonInfo, err := json.Marshal(info{identity, publicKey})
	if err != nil {
		w.WriteHeader(500)
	} else {
		fmt.Print(jsonInfo)
	}
}

// Show Blockchain
func Show(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(SBC.Show()))
	if err != nil {
		w.WriteHeader(500)
	}
	w.WriteHeader(200)
}

//
func ShowKeys(w http.ResponseWriter, r *http.Request) {
	compPubKeyMapJSON, _ := json.Marshal(compPubKeyMap)
	w.Write(compPubKeyMapJSON)
	userPubKeyMapJSON, _ := json.Marshal(userPubKeyMap)
	w.Write(userPubKeyMapJSON)
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

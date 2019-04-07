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

	go startTickin()
}

// Apply submits the application for a given user
func Apply(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	fmt.Print(string(body))
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
	cachemux.Lock()
	applicationCache[uid] = sub.Merit
	cachemux.Unlock()
}

func flushCache2BC() {
	acceptMpt := p1.MerklePatriciaTrie{}
	acceptMpt.Initial()
	applyMpt := p1.MerklePatriciaTrie{}
	applyMpt.Initial()

	cnt := 0

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
		cnt++
	}

	for k, v := range acceptanceCache {
		acceptMpt.Insert(k, string(v))
		delete(acceptanceCache, k)
		cnt++
	}

	if cnt > 0 {
		block := new(p2.Block)
		if SBC.Length() == 0 {
			block.Initial(1, "GENESIS", acceptMpt, applyMpt)
		} else {
			parentBlock, _ := SBC.Get(SBC.Length())
			fmt.Println(parentBlock)
			block.Initial(SBC.Length()+1, parentBlock[0].Header.Hash, acceptMpt, applyMpt)
		}

		SBC.Insert(*block)
	}
}

func startTickin() {
	for true {
		time.Sleep(10 * time.Second)
		flushCache2BC()
	}
}

// Fetch list of merits
func FetchMerits(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(SBC.ShowApplications()))
}

// Fetch list of acceptances
func FetchAcceptances(w http.ResponseWriter, r *http.Request) {
	jsonString, err := json.Marshal(SBC.ShowAcceptances())
	if err != nil {
		w.WriteHeader(500)
	}
	w.Write(jsonString)
}

func ViewCache(w http.ResponseWriter, r *http.Request) {
	applicationCacheJSON, _ := json.Marshal(applicationCache)
	w.Write([]byte("Application Cache"))
	w.Write(applicationCacheJSON)
	w.Write([]byte("\n"))
	w.Write([]byte("Acceptance Cache: "))
	acceptanceCacheJSON, _ := json.Marshal(acceptanceCache)
	w.Write(acceptanceCacheJSON)
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
	cachemux.Lock()
	defer cachemux.Unlock()
	vars := mux.Vars(r)
	// 1
	// 2
	company := vars["company"]
	uidTemp, err := strconv.Atoi(vars["uid"])

	fmt.Println(company)
	fmt.Println(uidTemp)
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
		w.Write((jsonInfo))
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

// ShowKeys show all the keys
func ShowKeys(w http.ResponseWriter, r *http.Request) {
	compPubKeyMapJSON, _ := json.Marshal(compPubKeyMap)
	w.Write([]byte("Company Public Keys: "))
	w.Write(compPubKeyMapJSON)
	w.Write([]byte("\n"))
	w.Write([]byte("Applicant Public Keys: "))
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

package p3

import (
	"../p3/data"
	"crypto/rsa"
	"net/http"
)

var SBC data.SyncBlockChain
var identityMap map[int32]data.Identity
var userPubKeyMap map[int32]rsa.PublicKey
var compPubKeyMap map[int32]rsa.PublicKey

// init will be executed before everything else.
// Some initialization will be done here.
func init() {
	SBC = data.NewBlockChain()
	identityMap = make(map[int32]data.Identity)
	userPubKeyMap = make(map[int32]rsa.PublicKey)
	compPubKeyMap = make(map[int32]rsa.PublicKey)
}

// Apply submits the application for a given user
func Apply(w http.ResponseWriter, r *http.Request) {

}

// Fetch list of job applications
func FetchApplications(w http.ResponseWriter, r *http.Request) {

}

// Register a business and their public key
func RegisterBusiness(w http.ResponseWriter, r *http.Request) {

}

// Accept a user
func Accept(w http.ResponseWriter, r *http.Request) {

}

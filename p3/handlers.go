package p3

import (
	"../p3/data"
	"net/http"
)

var DEFAULT_PORT int32 = 6686

var SBC data.SyncBlockChain

// init will be executed before everything else.
// Some initialization will be done here.
func init() {
	SBC = data.NewBlockChain()
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

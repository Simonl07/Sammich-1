package data

import (
	"sync"

	"../../p1"
	"../../p2"
)

type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

// NewBlockChain returns a new SyncBlockChain
func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: p2.NewBlockChain()}
}

// Get returns the list of blocks at a given height
func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	if height < 0 {
		return []p2.Block{}, false
	}
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Get(height), true
}

// GetBlock gets the block with the specific hash at the given height
func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	if height < 0 {
		return p2.Block{}, false
	}
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	for _, v := range sbc.bc.Chain[height] {
		if v.Header.Hash == hash {
			return v, true
		}
	}
	return p2.Block{}, false
}

// Insert inserts to the blockchain
func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	defer sbc.mux.Unlock()
}

// Length length of SBC
func (sbc *SyncBlockChain) Length() int32 {
	return sbc.bc.Length
}

// CheckParentHash adds the block if the parent hash exists
func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.CheckParentHash(insertBlock)
}

// UpdateEntireBlockChain decodes blockChainJson and adds the values to the blockchain
func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	sbc.bc.DecodeFromJson(blockChainJson)
}

// BlockChainToJson returns the json for the blockchain
func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.EncodeToJson()
}

// GenBlock generates a block at the next height
func (sbc *SyncBlockChain) GenBlock(acceptMpt p1.MerklePatriciaTrie, applyMpt p1.MerklePatriciaTrie) p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	blk, err := sbc.bc.GenBlock(acceptMpt, applyMpt)
	if err != nil {
		return p2.Block{}
	}
	// TODO: Check error
	return blk
}

// func (sbc *SyncBlockChain) AddToChain(submissions []Submission) {
// 	sbc.mux.Lock()
// 	defer sbc.mux.Unlock()
// 	sbc.bc.AddToChain(submissions)
// 	submissions = nil
// }

// Show returns a string representation of the underlying blockchain
func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}

package p1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"reflect"
	"strings"
)

// FlagValue is a struct that keeps track of the encoded prefix and the value
type FlagValue struct {
	encodedPrefix []uint8
	value         string
}

// Node is a generic node type that can represent Null, Branch, Ext, and Leaf types.
type Node struct {
	nodeType    int // 0: Null, 1: Branch, 2: Ext or Leaf
	branchValue [17]string
	flagValue   FlagValue
}

// MerklePatriciaTrie contains information about the MPT
type MerklePatriciaTrie struct {
	db     map[string]Node
	Root   string  `json:"root"`
	Values ValueDb `json:"valueDb"`
}

// ValueDb is a database that contains key values for use by JSON encoders.
type ValueDb struct {
	Db map[string]string `json:"mpt"`
}

// Get finds the value in the trie from a given key. It will return
// a blank string if the value could not be found. It will return an
// error if the mpt is not initialized.
func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	// Missing Key
	if key == "" {
		return "", nil
	}
	// Uninitialized Root
	if mpt.Root == "" || mpt.db == nil || mpt.db[mpt.Root].nodeType == 0 {
		return "", errors.New("uninitialized trie")
	}
	// recursively find item
	return mpt.recurseGet(asciiToHexArray([]uint8(key)), mpt.Root), nil
}

// recurseGet is a recursive helper for Get to hunt for the given key.
// It returns the found value of "" if not found.
func (mpt *MerklePatriciaTrie) recurseGet(hexKey []uint8, currHash string) string {
	currNode := mpt.db[currHash]
	if currNode.nodeType == 2 { // leaf or ext
		currDecPrefix := compactDecode(currNode.flagValue.encodedPrefix)
		increment := similar(currDecPrefix, hexKey)
		if increment != len(currDecPrefix) {
			return ""
		}
		if isExtNode(currNode.flagValue.encodedPrefix) { // ext
			return mpt.recurseGet(hexKey[increment:], currNode.flagValue.value)
		} else { // leaf
			return currNode.flagValue.value
		}
	} else if currNode.nodeType == 1 { // branch
		if len(hexKey) == 0 {
			return currNode.branchValue[16]
		}
		for _, v := range currNode.branchValue[:16] {
			if v != "" {
				result := mpt.recurseGet(hexKey[1:], v)
				if result != "" {
					return result
				}
			}
		}
	}
	return ""
}

// Insert finds the correct location in the trie and inserts the value.
func (mpt *MerklePatriciaTrie) Insert(key string, newValue string) {
	// key and value should NOT be empty
	if len(key) == 0 || len(newValue) == 0 {
		return
	}
	// If this trie has no Root and is new
	hexKey := asciiToHexArray([]uint8(key))
	// If unitialied Root, initialize it
	if mpt.Root == "" || mpt.db == nil || mpt.db[mpt.Root].nodeType == 0 {
		node := Node{2, [17]string{}, FlagValue{compactEncode(append(hexKey, 16)), newValue}}
		hash := node.hashNode()
		mpt.db = make(map[string]Node)
		mpt.Values.Db = make(map[string]string)
		mpt.db[hash] = node
		mpt.Values.Db[key] = newValue
		mpt.Root = hash
		return
	}
	// Recursively find correct slot to insert into
	currHash := mpt.Root
	mpt.Root = mpt.recurseInsert(currHash, hexKey, newValue)
	mpt.Values.Db[key] = newValue
}

// refreshHash updates the old hash with the new hash.
// It returns the new hash value.
func (mpt *MerklePatriciaTrie) refreshHash(oldHash string, newNode Node) string {
	newHash := newNode.hashNode()
	delete(mpt.db, oldHash)
	mpt.db[newHash] = newNode
	return newHash
}

// similar finds the number of similar items from the beginning of each uint8 array.
// It returns the number continual similarities between the two arrays.
func similar(arr1 []uint8, arr2 []uint8) int {
	same := 0
	for i, v := range arr1 {
		if i < len(arr2) && v == arr2[i] {
			same++
		} else {
			break
		}
	}
	return same
}

// recurseInsert is a helper function for Insert to help find the location they key should be inserted
// and rehash up the trie. It returns the highest level hash in the current recursive step.
func (mpt *MerklePatriciaTrie) recurseInsert(currHash string, currHexKey []uint8, value string) string {
	currNode := mpt.db[currHash]
	if currNode.nodeType == 1 { // Case 1: Current node is a branch
		if len(currHexKey) == 0 { // Case 1a: Leaf value should be inserted to branch
			currNode.branchValue[16] = value
			return mpt.refreshHash(currHash, currNode)
		} else {
			nextHash := currNode.branchValue[currHexKey[0]]
			var hash = ""
			if nextHash == "" { // Case 1b: Curr is branch and array pos is empty
				// Create leaf
				node := Node{2, [17]string{}, FlagValue{compactEncode(append(currHexKey[1:], 16)), value}}
				hash = node.hashNode()
				mpt.db[hash] = node
			} else { // Case 1c: Curr is branch and array pos is full
				// Recursive call, get next newHash (should be extension or branch)
				hash = mpt.recurseInsert(nextHash, currHexKey[1:], value)
			}
			currNode.branchValue[currHexKey[0]] = hash
			// Rehash the current node
			return mpt.refreshHash(currHash, currNode)
		}
	} else if currNode.nodeType == 2 { // Case 2: Current is a extension or leaf
		decodedPrefix := compactDecode(currNode.flagValue.encodedPrefix)
		same := similar(decodedPrefix, currHexKey)
		if isExtNode(currNode.flagValue.encodedPrefix) { // Case 2a: Current is extension
			if same > 0 {
				if same == len(decodedPrefix) { // Case 2aa: ext matches beginning of key. recurse further
					newNextHash := mpt.recurseInsert(currNode.flagValue.value, currHexKey[len(currNode.flagValue.encodedPrefix):], value)
					// newNextHash should ALWAYS refer to a branch node
					currNode.flagValue.value = newNextHash
					return mpt.refreshHash(currHash, currNode)
				} else if same == len(decodedPrefix)-1 { // Case 2ab: ext~> => ext->branch->leaf,~>
					// Create new leaf
					node := Node{2, [17]string{}, FlagValue{compactEncode(append(currHexKey[same+1:], 16)), value}}
					hash := node.hashNode()
					mpt.db[hash] = node
					// Create new branch
					branchNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
					branchNode.branchValue[decodedPrefix[same]] = currNode.flagValue.value
					branchNode.branchValue[currHexKey[same]] = hash
					branchHash := branchNode.hashNode()
					mpt.db[branchHash] = branchNode
					// Edit current extension
					currNode.flagValue.encodedPrefix = compactEncode(decodedPrefix[:same])
					currNode.flagValue.value = branchHash
					return mpt.refreshHash(currHash, currNode)
				} else if len(currHexKey) == same {
					// Create new extension
					// currNode.flagValue.value should ONLY be a branch node
					extNode := Node{2, [17]string{}, FlagValue{compactEncode(decodedPrefix[same+1:]), currNode.flagValue.value}}
					extHash := extNode.hashNode()
					delete(mpt.db, currHash)
					mpt.db[extHash] = extNode
					// Create new branch
					branchNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
					branchNode.branchValue[decodedPrefix[same]] = extHash
					branchNode.branchValue[16] = value
					branchHash := branchNode.hashNode()
					mpt.db[branchHash] = branchNode
					// Edit current extension
					currNode.flagValue.encodedPrefix = compactEncode(decodedPrefix[:same])
					currNode.flagValue.value = branchHash
					newCurrHash := currNode.hashNode()
					mpt.db[newCurrHash] = currNode
					return newCurrHash
				} else {
					// Create new extension
					// currNode.flagValue.value should ONLY be a branch node
					extNode := Node{2, [17]string{}, FlagValue{compactEncode(decodedPrefix[same+1:]), currNode.flagValue.value}}
					extHash := extNode.hashNode()
					delete(mpt.db, currHash)
					mpt.db[extHash] = extNode
					// Create new leaf
					node := Node{2, [17]string{}, FlagValue{compactEncode(append(currHexKey[same+1:], 16)), value}}
					hash := node.hashNode()
					mpt.db[hash] = node
					// Create new branch
					branchNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
					branchNode.branchValue[decodedPrefix[same]] = extHash
					branchNode.branchValue[currHexKey[same]] = hash
					branchHash := branchNode.hashNode()
					mpt.db[branchHash] = branchNode
					// Edit current extension
					currNode.flagValue.encodedPrefix = compactEncode(decodedPrefix[:same])
					currNode.flagValue.value = branchHash
					newCurrHash := currNode.hashNode()
					mpt.db[newCurrHash] = currNode
					return newCurrHash
				}
			} else {
				if len(currHexKey) == 0 {
					// Create new branch
					newBNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
					newBNode.branchValue[16] = value
					branchPos := decodedPrefix[0]
					if len(decodedPrefix) <= 1 { // Should be at least 1
						// If ext prefix length is 1, then remove the ext
						delete(mpt.db, currHash)
						newBNode.branchValue[branchPos] = currNode.flagValue.value
						// rehash branch
						bHash := newBNode.hashNode()
						mpt.db[bHash] = newBNode
						return bHash
					} else {
						// If ext prefix length is more than 1, then modify the ext
						currNode.flagValue.encodedPrefix = compactEncode(decodedPrefix[1:])
						newHash := mpt.refreshHash(currHash, currNode)
						newBNode.branchValue[branchPos] = newHash
						// rehash branch
						bHash := newBNode.hashNode()
						mpt.db[bHash] = newBNode
						return bHash
					}
				} else {
					// We know len(currHexKey) > 0 and same == 0 at this point
					// Create new leaf
					node := Node{2, [17]string{}, FlagValue{compactEncode(append(currHexKey[1:], 16)), value}}
					hash := node.hashNode()
					mpt.db[hash] = node
					// Create new branch
					newBNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
					newBNode.branchValue[currHexKey[0]] = hash
					branchPos := decodedPrefix[0]
					if len(decodedPrefix) <= 1 {
						// If curr ext length is 1, then remove it.
						delete(mpt.db, currHash)
						newBNode.branchValue[branchPos] = currNode.flagValue.value
						// rehash branch node
						bHash := newBNode.hashNode()
						mpt.db[bHash] = newBNode
						return bHash
					} else {
						// If curr ext length is greater than 1, then edit the ext
						currNode.flagValue.encodedPrefix = compactEncode(decodedPrefix[1:])
						newHash := mpt.refreshHash(currHash, currNode)
						newBNode.branchValue[branchPos] = newHash
						// rehash the branch node
						bHash := newBNode.hashNode()
						mpt.db[bHash] = newBNode
						return bHash
					}
				}
			}
		} else { // Case 2b: Current is leaf
			// Case 2ba: Leaf value will be replaced
			if same == len(currHexKey) && same == len(decodedPrefix) {
				currNode.flagValue.value = value
				return mpt.refreshHash(currHash, currNode)
			} else if len(currHexKey) == 0 {
				// Create new branch
				newBNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
				newBNode.branchValue[16] = value
				//
				branchPos := decodedPrefix[0]
				currNode.flagValue.encodedPrefix = compactEncode(append(decodedPrefix[1:], 16))
				newHash := mpt.refreshHash(currHash, currNode)
				newBNode.branchValue[branchPos] = newHash
				//
				bHash := newBNode.hashNode()
				mpt.db[bHash] = newBNode
				return bHash
			} else if len(decodedPrefix) == 0 {
				// Create new leaf
				node := Node{2, [17]string{}, FlagValue{compactEncode(append(currHexKey[1:], 16)), value}}
				hash := node.hashNode()
				mpt.db[hash] = node
				// Create new branch
				newBNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
				newBNode.branchValue[16] = currNode.flagValue.value
				delete(mpt.db, currHash)
				newBNode.branchValue[currHexKey[0]] = hash
				//
				bHash := newBNode.hashNode()
				mpt.db[bHash] = newBNode
				return bHash
			} else if len(decodedPrefix) == same {
				// Create new branch
				newBNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
				newBNode.branchValue[16] = currNode.flagValue.value
				//
				branchPos := currHexKey[same]
				currNode.flagValue.encodedPrefix = compactEncode(append(currHexKey[same+1:], 16))
				currNode.flagValue.value = value
				newHash := mpt.refreshHash(currHash, currNode)
				newBNode.branchValue[branchPos] = newHash
				//
				bHash := newBNode.hashNode()
				mpt.db[bHash] = newBNode
				// Create new extension
				extNode := Node{2, [17]string{}, FlagValue{compactEncode(currHexKey[:same]), bHash}}
				extHash := extNode.hashNode()
				mpt.db[extHash] = extNode
				return extHash
			} else if len(currHexKey) == same {
				// Create new branch
				newBNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
				newBNode.branchValue[16] = value
				//
				branchPos := decodedPrefix[same]
				currNode.flagValue.encodedPrefix = compactEncode(append(decodedPrefix[same+1:], 16))
				newHash := mpt.refreshHash(currHash, currNode)
				newBNode.branchValue[branchPos] = newHash
				//
				bHash := newBNode.hashNode()
				mpt.db[bHash] = newBNode
				// Create new extension
				extNode := Node{2, [17]string{}, FlagValue{compactEncode(currHexKey[:same]), bHash}}
				extHash := extNode.hashNode()
				mpt.db[extHash] = extNode
				return extHash
			} else if same > 0 {
				// Case 2bb: Current is leaf and the currNode path overlaps with the insert path.
				//           This means that we will use an extension.
				// Create new leaf
				node := Node{2, [17]string{}, FlagValue{compactEncode(append(currHexKey[same+1:], 16)), value}}
				hash := node.hashNode()
				mpt.db[hash] = node
				// Create new branch
				newBNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
				newBNode.branchValue[currHexKey[same]] = hash
				//
				branchPos := decodedPrefix[same]
				currNode.flagValue.encodedPrefix = compactEncode(append(decodedPrefix[same+1:], 16))
				newHash := mpt.refreshHash(currHash, currNode)
				newBNode.branchValue[branchPos] = newHash
				//
				bHash := newBNode.hashNode()
				mpt.db[bHash] = newBNode
				// Create new extension
				extNode := Node{2, [17]string{}, FlagValue{compactEncode(currHexKey[:same]), bHash}}
				extHash := extNode.hashNode()
				mpt.db[extHash] = extNode
				return extHash
			} else {
				// Case 2bc: Current is leaf and there is no overlap in path. So we will not add an extension.
				if len(currHexKey) == 0 {
					currNode.flagValue.value = value
					return mpt.refreshHash(currHash, currNode)
				}
				// Create new leaf
				node := Node{2, [17]string{}, FlagValue{compactEncode(append(currHexKey[1:], 16)), value}}
				hash := node.hashNode()
				mpt.db[hash] = node
				// Create new branch
				newBNode := Node{1, [17]string{}, FlagValue{[]uint8{}, ""}}
				newBNode.branchValue[currHexKey[same]] = hash
				//
				if len(decodedPrefix) == 0 {
					newBNode.branchValue[16] = currNode.flagValue.value
					delete(mpt.db, currHash)
				} else {
					branchPos := decodedPrefix[same]
					currNode.flagValue.encodedPrefix = compactEncode(append(decodedPrefix[1:], 16))
					newHash := mpt.refreshHash(currHash, currNode)
					newBNode.branchValue[branchPos] = newHash
				}
				//
				bHash := newBNode.hashNode()
				mpt.db[bHash] = newBNode
				return bHash
			}
		}
	}
	return "" // This should NEVER happen
}

// Delete removes the given key from the MPT if it exists.
// Delete will return "" if the key does not exist. Otherwise, it will return the value that was removed. If the key
// is missing or if the tree is uninitialized, and error will be thrown.
func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	if len(key) == 0 {
		return "", errors.New("missing key")
	}

	// If this trie has no Root and is new
	hexKey := asciiToHexArray([]uint8(key))
	if mpt.Root == "" || mpt.db == nil || mpt.db[mpt.Root].nodeType == 0 {
		return "", errors.New("uninitialized trie")
	}

	item, nodeHash, remove := mpt.recurseDelete(hexKey, mpt.Root)
	if remove {
		mpt.Root = ""
		mpt.db = nil
	} else {
		mpt.Root = nodeHash
	}
	delete(mpt.Values.Db, key)
	return item, nil
}

// recursiveDelete is the helper function for Delete.
// The function recursively searches and updates the trie hashes for an item. It returns the value that was deleted,
// the highest level hash in the recursive step, and whether the value should be removed. If the value should not
// be remove then just rehash.
func (mpt *MerklePatriciaTrie) recurseDelete(hexKey []uint8, currHash string) (string, string, bool) {
	currNode := mpt.db[currHash]
	if currNode.nodeType == 1 { // branch
		if len(hexKey) == 0 {
			val := currNode.branchValue[16]
			currNode.branchValue[16] = ""
			filled := 0
			var idx = -1
			var lastHash = ""
			for i, v := range currNode.branchValue[:len(currNode.branchValue)-1] {
				if v != "" {
					filled++
					lastHash = v
					idx = i
				}
			}
			if filled == 1 {
				lastNode := mpt.db[lastHash]
				if lastNode.nodeType == 2 && !isExtNode(lastNode.flagValue.encodedPrefix) {
					// leaf
					lastNode.flagValue.encodedPrefix = compactEncode(append([]uint8{uint8(idx)}, append(compactDecode(lastNode.flagValue.encodedPrefix), 16)...))
				} else {
					lastNode.flagValue.encodedPrefix = compactEncode(append([]uint8{uint8(idx)}, compactDecode(lastNode.flagValue.encodedPrefix)...))
				}
				delete(mpt.db, currHash)
				return val, mpt.refreshHash(lastHash, lastNode), false
			} else {
				return val, mpt.refreshHash(currHash, currNode), false
			}
		}
		if currNode.branchValue[hexKey[0]] == "" {
			return "", currHash, false
		}
		item, nodeHash, remove := mpt.recurseDelete(hexKey[1:], currNode.branchValue[hexKey[0]])
		if item == "" {
			return "", currHash, false
		}
		node := mpt.db[nodeHash]
		if node.nodeType == 2 {
			if isExtNode(node.flagValue.encodedPrefix) || !remove {
				currNode.branchValue[hexKey[0]] = node.hashNode()
				return item, mpt.refreshHash(currHash, currNode), false
			} else {
				filled := 0
				var lastHash = ""
				var idx = -1
				for i, v := range currNode.branchValue[:len(currNode.branchValue)-1] {
					if v != "" {
						filled++
						if i != int(hexKey[0]) {
							lastHash = v
							idx = i
						}
					}
				}
				if filled > 2 {
					currNode.branchValue[hexKey[0]] = ""
					return item, mpt.refreshHash(currHash, currNode), false
				} else if currNode.branchValue[16] != "" {
					node := Node{2, [17]string{}, FlagValue{compactEncode(append([]uint8{uint8(idx)}, 16)), currNode.branchValue[16]}}
					hash := node.hashNode()
					mpt.db[hash] = node
					return item, hash, false
				} else if filled > 1 {
					lastNode := mpt.db[lastHash]
					if lastNode.nodeType == 2 {
						if !isExtNode(lastNode.flagValue.encodedPrefix) {
							// leaf
							lastNode.flagValue.encodedPrefix = compactEncode(append([]uint8{uint8(idx)}, append(compactDecode(lastNode.flagValue.encodedPrefix), 16)...))
						} else {
							lastNode.flagValue.encodedPrefix = compactEncode(append([]uint8{uint8(idx)}, compactDecode(lastNode.flagValue.encodedPrefix)...))
						}
						delete(mpt.db, currHash)
						return item, mpt.refreshHash(lastHash, lastNode), false
					} else if lastNode.nodeType == 1 {
						// Convert branch to extension
						currNode.nodeType = 2
						currNode.flagValue.value = lastHash
						currNode.branchValue = [17]string{}
						currNode.flagValue.encodedPrefix = compactEncode([]uint8{uint8(idx)})
						return item, mpt.refreshHash(currHash, currNode), false
					}
				} else {
					return item, currHash, false
				}
			}
		} else if node.nodeType == 1 {
			currNode.branchValue[hexKey[0]] = nodeHash
			return item, mpt.refreshHash(currHash, currNode), false
		}
		return item, currHash, false
	} else if currNode.nodeType == 2 { // leaf or ext
		if isExtNode(currNode.flagValue.encodedPrefix) { // ext
			decodedPrefix := compactDecode(currNode.flagValue.encodedPrefix)
			if len(hexKey) < len(decodedPrefix) {
				return "", currHash, false
			}
			for i, v := range decodedPrefix {
				if v != hexKey[i] {
					return "", currHash, false
				}
			}
			item, nodeHash, _ := mpt.recurseDelete(hexKey[len(decodedPrefix):], currNode.flagValue.value)
			if item == "" {
				return "", currHash, false
			}
			node := mpt.db[nodeHash]
			if node.nodeType == 2 {
				if isExtNode(node.flagValue.encodedPrefix) {
					currNode.flagValue.encodedPrefix = compactEncode(
						append(decodedPrefix, compactDecode(node.flagValue.encodedPrefix)...))
					currNode.flagValue.value = node.flagValue.value
					return item, mpt.refreshHash(currHash, currNode), false
				} else {
					node.flagValue.encodedPrefix = compactEncode(
						append(append(decodedPrefix, compactDecode(node.flagValue.encodedPrefix)...), 16))
					delete(mpt.db, currHash)
					mpt.db[node.hashNode()] = node
					return item, nodeHash, false
				}
			} else if node.nodeType == 1 {
				currNode.flagValue.value = nodeHash
				return item, mpt.refreshHash(currHash, currNode), false
			}
			return item, currHash, false
		} else { // leaf
			decodedPrefix := compactDecode(currNode.flagValue.encodedPrefix)
			if len(hexKey) != len(decodedPrefix) {
				return "", currHash, false
			}
			for i, v := range decodedPrefix {
				if v != hexKey[i] {
					return "", currHash, false
				}
			}
			return currNode.flagValue.value, currHash, true
		}
	}
	return "", currHash, false
}

// compactEncode encodes an hexArray to an ASCII array with the specified attributes in the link below.
// This code is based on https://github.com/ethereum/wiki/wiki/Patricia-Tree#main-specification-merkle-patricia-trie
// This function returns the encoded array.
func compactEncode(hexArray []uint8) []uint8 {
	term := hexArray[len(hexArray)-1] == 16
	slice := hexArray[:]
	if term {
		slice = hexArray[:len(hexArray)-1]
	}

	oddLen := len(slice)%2 == 1

	var flags uint8 = 0
	if term {
		flags += 2
	}
	if oddLen {
		flags += 1
	}

	if oddLen {
		slice = append([]uint8{flags}, slice...)
	} else {
		slice = append([]uint8{flags, 0}, slice...)
	}

	encodedArr := make([]uint8, len(slice)/2)
	for i := 0; i < len(encodedArr); i++ {
		encodedArr[i] = (16 * slice[2*i]) + slice[(2*i)+1]
	}
	return encodedArr
}

// compactDecode decodes the encoded ASCII array to a hex array.
// This functions returns the decoded array.
func compactDecode(encodedArr []uint8) []uint8 {
	expandedArr := asciiToHexArray(encodedArr)

	trimmedArr := expandedArr[1:]
	if expandedArr[0] == 0 || expandedArr[0] == 2 {
		trimmedArr = trimmedArr[1:]
	}

	return trimmedArr
}

// asciiToHexArray converts from an ASCII array to hex array.
// This function returns the converted hex array.
func asciiToHexArray(arr []uint8) []uint8 {
	expandedArr := make([]uint8, len(arr)*2)
	for i, v := range arr {
		expandedArr[i*2] = v / 16
		expandedArr[(i*2)+1] = v % 16
	}
	return expandedArr
}

// testCompactEncode is a simple test function for compactEncode and compactDecode.
func testCompactEncode() {
	fmt.Println(reflect.DeepEqual(compactDecode(compactEncode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compactDecode(compactEncode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compactDecode(compactEncode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compactDecode(compactEncode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

// hashNode hashes a node with HashStart and HashEnd at either end for clarity.
// This function returns the hash string.
// TODO: Remove HashStart and HashEnd
func (node *Node) hashNode() string {
	var str string
	switch node.nodeType {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branchValue {
			str += v
		}
	case 2:
		str = node.flagValue.value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

// String returns the string representation of the Node.
func (node *Node) String() string {
	str := "empty string"
	switch node.nodeType {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.branchValue[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.branchValue[16])
	case 2:
		encodedPrefix := node.flagValue.encodedPrefix
		nodeName := "Leaf"
		if isExtNode(encodedPrefix) {
			nodeName = "Ext"
		}
		oriPrefix := strings.Replace(fmt.Sprint(compactDecode(encodedPrefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", nodeName, oriPrefix, node.flagValue.value)
	}
	return str
}

// nodeToString takes a node and converts it to a string. This string is returned.
func nodeToString(node Node) string {
	return node.String()
}

// Initial functions like a simple constructor for the MerklePatriciaTrie.
func (mpt *MerklePatriciaTrie) Initial() {
	mpt.db = make(map[string]Node)
}

// Clone mpt (make deep copy)
func (mpt *MerklePatriciaTrie) Clone() MerklePatriciaTrie {
	mpt2 := MerklePatriciaTrie{}
	mpt2.Initial()
	for k, v := range mpt2.Values.Db {
		mpt2.Insert(k, v)
	}
	return mpt2
}

// isExtNode tests if the encoded array is an extension node. The boolean evaluated is returned.
func isExtNode(encodedArr []uint8) bool {
	return encodedArr[0]/16 < 2
}

// TestCompact invokes the testCompactEncode function to test the encoding and decoding functions.
func TestCompact() {
	testCompactEncode()
}

// String converts a MerklePatriciaTrie to a string representation. This string is returned.
func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.Root)
	for hash := range mpt.db {
		content += fmt.Sprintf("%s: %s\n", hash, nodeToString(mpt.db[hash]))
	}
	return content
}

// OrderNodes orders the nodes in a string representation. The string is returned.
func (mpt *MerklePatriciaTrie) OrderNodes() string {
	rawContent := mpt.String()
	content := strings.Split(rawContent, "\n")
	rootHash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{rootHash}
	i := -1
	rs := ""
	curHash := ""
	for len(queue) != 0 {
		lastIndex := len(queue) - 1
		curHash, queue = queue[lastIndex], queue[:lastIndex]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+curHash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+curHash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}

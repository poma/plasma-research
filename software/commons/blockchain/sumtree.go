package blockchain

import (
	a "../alias"
	"encoding/binary"
	"sort"
)
import u "../utils"
import s "../../plasmautils/slice"

type SumTreeNode struct {
	// We use 24 bit
	Length uint32
	Hash   a.Uint160

	//isLeft  bool
	//isRight bool

	Left   *SumTreeNode
	Right  *SumTreeNode
	Parent *SumTreeNode
}

// Use this first when assemble blocks
func PrepareLeaves(transactions []Transaction) []SumTreeNode {

	slice2transactions := map[s.Slice]*Transaction{}

	var slices []s.Slice
	for _, t := range transactions {
		for _, input := range t.Inputs {
			slices = append(slices, input.Slice)
			slice2transactions[input.Slice] = &t
		}
	}

	sort.Slice(slices, func(i, j int) bool {
		return slices[i].Begin < slices[j].Begin
	})

	var leafs []SumTreeNode
	for _, slice := range slices {
		leaf := SumTreeNode{
			Hash:   slice2transactions[slice].GetHash(),
			Length: slice.End - slice.Begin,
		}

		leafs = append(leafs, leaf)
	}

	// TODO: look do we need padding at that moment
	if len(leafs)%2 == 1 {
		emptyLeaf := SumTreeNode{
			Hash:   u.Keccak160([]byte{}), // Hash from empty byte array
			Length: 0,
		}
		leafs = append(leafs, emptyLeaf)
	}

	return leafs
}

type SumMerkleTree struct {
	Root  SumTreeNode
	Leafs []SumTreeNode
}

// TODO: check that way is compatible with soidity
// Uint to bytes
func u2b(value uint32) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b, value)
	return b
}

func concatAndHash(left SumTreeNode, right SumTreeNode) a.Uint160 {
	l1, l2 := left.Length, right.Length
	h1, h2 := left.Hash, right.Hash

	d1 := append(u2b(l1), u2b(l2)...)
	d2 := append(h1, h2...)
	return u.Keccak160(append(d1, d2...))
}

func NewSumMerkleTree(leafs []SumTreeNode) *SumMerkleTree {
	var tree SumMerkleTree
	tree.Leafs = leafs

	var buckets = tree.Leafs

	// At the end we assign new layer to buckets, so stop when ever we can't merge anymore
	for len(buckets) != 1 {
		// next layer
		var newBuckets []SumTreeNode

		for len(buckets) > 0 {
			if len(buckets) >= 2 {

				// deque pair from the head
				left, right := buckets[0], buckets[1]
				buckets = buckets[2:]

				length := left.Length + right.Length
				hash := concatAndHash(left, right)

				node := SumTreeNode{
					Hash:   hash,
					Length: length,
				}

				left.Parent = &node
				right.Parent = &node

				left.Right = &right
				right.Left = &left

				newBuckets = append(newBuckets, node)

			} else {
				// Pop the last one - goes to next layer as it is
				newBuckets = append(newBuckets, buckets[0])
				buckets = []SumTreeNode{}
			}
		}

		buckets = newBuckets
	}

	tree.Root = buckets[0]

	return &tree
}

func (tree *SumMerkleTree) GetProof(leafIndex uint32) []byte {
	return []byte{0x0}
	// return tree.Root.Length, tree.Root.Hash
}

//func (tree *SumMerkleTree) GetRoot() (length uint32, hash a.Uint160) {
//	return tree.Root.Length, tree.Root.Hash
//}

func (tree *SumMerkleTree) GetRoot() SumTreeNode {
	return tree.Root
}

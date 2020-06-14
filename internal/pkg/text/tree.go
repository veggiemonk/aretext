package text

import (
	"errors"
	"io"
)

// text.Tree is a data structure for representing UTF-8 text.
// It supports efficient insertions, deletions, and lookup by character offset and line number.
// It is inspired by two papers:
// Boehm, H. J., Atkinson, R., & Plass, M. (1995). Ropes: an alternative to strings. Software: Practice and Experience, 25(12), 1315-1330.
// Rao, J., & Ross, K. A. (2000, May). Making B+-trees cache conscious in main memory. In Proceedings of the 2000 ACM SIGMOD international conference on Management of data (pp. 475-486).
// Like a rope, the tree maintains character counts at each level to efficiently locate a character at a given offset.
// To use the CPU cache efficiently, all children of a node are pre-allocated in a group (what the Rao & Ross paper calls a "full" cache-sensitive B+ tree),
// and the parent uses offsets within the node group to identify child nodes.
// All nodes are carefully designed to fit as much data as possible within a 64-byte cache line.
type Tree struct {
	root      nodeGroup
	validator *Validator
}

// NewTree returns a tree representing an empty string.
func NewTree() *Tree {
	validator := NewValidator()
	root := &innerNodeGroup{numNodes: 1}
	root.nodes[0].child = &leafNodeGroup{numNodes: 1}
	root.nodes[0].numKeys = 1
	return &Tree{root, validator}
}

// NewTreeFromReader creates a new Tree from a reader that produces UTF-8 text.
// This is more efficient than inserting the bytes into an empty tree.
// Returns an error if the bytes are invalid UTF-8.
func NewTreeFromReader(r io.Reader) (*Tree, error) {
	validator := NewValidator()
	leafGroups, err := bulkLoadIntoLeaves(r, validator)
	if err != nil {
		return nil, err
	}
	root := buildInnerNodesFromLeaves(leafGroups)
	return &Tree{root, validator}, nil
}

func bulkLoadIntoLeaves(r io.Reader, v *Validator) ([]nodeGroup, error) {
	leafGroups := make([]nodeGroup, 0, 1)
	currentGroup := &leafNodeGroup{numNodes: 1}
	currentNode := &currentGroup.nodes[0]
	leafGroups = append(leafGroups, currentGroup)

	var buf [1024]byte
	for {
		n, err := r.Read(buf[:])
		if err != nil && err != io.EOF {
			return nil, err
		}

		if n == 0 {
			break
		}

		if !v.ValidateBytes(buf[:n]) {
			return nil, errors.New("invalid UTF-8")
		}

		for i := 0; i < n; i++ {
			if currentNode.numBytes == maxBytesPerLeaf {
				if currentGroup.numNodes < uint64(maxNodesPerGroup) {
					currentNode = &currentGroup.nodes[currentGroup.numNodes]
					currentGroup.numNodes++
				} else {
					newGroup := &leafNodeGroup{numNodes: 1}
					leafGroups = append(leafGroups, newGroup)
					newGroup.prev = currentGroup
					currentGroup.next = newGroup
					currentGroup = newGroup
					currentNode = &currentGroup.nodes[0]
				}
			}

			currentNode.textBytes[currentNode.numBytes] = buf[i]
			currentNode.numBytes++
		}
	}

	if !v.ValidateEnd() {
		return nil, errors.New("invalid UTF-8")
	}

	return leafGroups, nil
}

func buildInnerNodesFromLeaves(leafGroups []nodeGroup) nodeGroup {
	childGroups := leafGroups

	for {
		parentGroups := make([]nodeGroup, 0, len(childGroups)/int(maxNodesPerGroup)+1)
		currentGroup := &innerNodeGroup{}
		parentGroups = append(parentGroups, currentGroup)

		for _, cg := range childGroups {
			if currentGroup.numNodes == uint64(maxNodesPerGroup) {
				newGroup := &innerNodeGroup{}
				parentGroups = append(parentGroups, newGroup)
				currentGroup = newGroup
			}

			innerNode := &currentGroup.nodes[currentGroup.numNodes]
			innerNode.child = cg
			for i, key := range cg.keys() {
				innerNode.keys[i] = key
				innerNode.numKeys++
			}
			currentGroup.numNodes++
		}

		if len(parentGroups) == 1 && currentGroup.numNodes == 1 {
			return parentGroups[0]
		}

		childGroups = parentGroups
	}
}

// CursorAtPosition returns a cursor starting at the UTF-8 character at the specified position (0-indexed).
// If the position is past the end of the text, the returned cursor will read zero bytes.
func (t *Tree) CursorAtPosition(charPos uint64) *Cursor {
	return t.root.cursorAtPosition(0, charPos)
}

// text.Cursor reads UTF-8 bytes from a text.Tree.
// It implements io.Reader.
// text.Tree is NOT thread-safe, so reading from a tree while modifying it is undefined behavior!
type Cursor struct {
	group          *leafNodeGroup
	nodeIdx        byte
	textByteOffset byte
}

func (c *Cursor) Read(b []byte) (int, error) {
	i := 0
	for {
		if i == len(b) {
			return i, nil
		}

		if c.group == nil {
			return i, io.EOF
		}

		node := &c.group.nodes[c.nodeIdx]
		bytesWritten := copy(b[i:], node.textBytes[c.textByteOffset:node.numBytes])
		c.textByteOffset += byte(bytesWritten) // conversion is safe b/c maxBytesPerLeaf < 256
		i += bytesWritten

		if c.textByteOffset == node.numBytes {
			c.nodeIdx++
			c.textByteOffset = 0
		}

		if uint64(c.nodeIdx) == c.group.numNodes {
			c.group = c.group.next
			c.nodeIdx = 0
			c.textByteOffset = 0
		}
	}

	return 0, nil
}

const maxKeysPerNode = byte(64)
const maxNodesPerGroup = maxKeysPerNode
const maxBytesPerLeaf = 63

// nodeGroup is either an inner node group or a leaf node group.
type nodeGroup interface {
	keys() []indexKey
	cursorAtPosition(nodeIdx byte, charPos uint64) *Cursor
}

// indexKey is used to navigate from an inner node to the child node containing a particular line or character offset.
type indexKey struct {

	// Number of UTF-8 characters in a subtree.
	// If a multi-byte UTF-8 character is split between multiple leaf nodes,
	// it is counted by the node containing the first byte.
	numChars uint64

	// Number of UTF-8 characters in a subtree.
	// If a multi-byte UTF-8 character is split between multiple leaf nodes,
	// it is counted by the node containing the first byte.
	numLines uint64
}

// innerNodeGroup is a group of inner nodes referenced by a parent inner node.
type innerNodeGroup struct {
	numNodes uint64
	nodes    [maxNodesPerGroup]innerNode
}

func (g *innerNodeGroup) keys() []indexKey {
	keys := make([]indexKey, g.numNodes)
	for i := uint64(0); i < g.numNodes; i++ {
		keys[i] = g.nodes[i].key()
	}
	return keys
}

func (g *innerNodeGroup) cursorAtPosition(nodeIdx byte, charPos uint64) *Cursor {
	return g.nodes[nodeIdx].cursorAtPosition(charPos)
}

// innerNode is used to navigate to the leaf node containing a character offset or line number.
//
// +-----------------------------------------+
// | child | numKeys |  padding   | keys[64] |
// +-----------------------------------------+
//     8        1          7          1024     = 1032 bytes
//
type innerNode struct {
	child   nodeGroup
	numKeys byte

	// Each key corresponds to a node in the child group.
	keys [maxKeysPerNode]indexKey
}

func (n *innerNode) key() indexKey {
	nodeKey := indexKey{}
	for i := byte(0); i < n.numKeys; i++ {
		key := n.keys[i]
		nodeKey.numChars += key.numChars
		nodeKey.numLines += key.numLines
	}
	return nodeKey
}

func (n *innerNode) cursorAtPosition(charPos uint64) *Cursor {
	c := uint64(0)

	for i := byte(0); i < n.numKeys-1; i++ {
		nc := n.keys[i].numChars
		if charPos < c+nc {
			return n.child.cursorAtPosition(i, charPos-c)
		}
		c += nc
	}

	return n.child.cursorAtPosition(n.numKeys-1, charPos-c)
}

// leafNodeGroup is a group of leaf nodes referenced by an inner node.
// These form a doubly-linked list so a cursor can scan the text efficiently.
type leafNodeGroup struct {
	prev     *leafNodeGroup
	next     *leafNodeGroup
	numNodes uint64
	nodes    [maxNodesPerGroup]leafNode
}

func (g *leafNodeGroup) keys() []indexKey {
	keys := make([]indexKey, g.numNodes)
	for i := uint64(0); i < g.numNodes; i++ {
		keys[i] = g.nodes[i].key()
	}
	return keys
}

func (g *leafNodeGroup) cursorAtPosition(nodeIdx byte, charPos uint64) *Cursor {
	textByteOffset := g.nodes[nodeIdx].byteOffsetForPosition(charPos)
	return &Cursor{
		group:          g,
		nodeIdx:        nodeIdx,
		textByteOffset: textByteOffset,
	}
}

// leafNode is a node that stores UTF-8 text as a byte array.
//
// Multi-byte UTF-8 characters may be split between adjacent leaf nodes.
//
// +---------------------------------+
// |   numBytes  |   textBytes[63]   |
// +---------------------------------+
//        1               63          = 64 bytes
//
type leafNode struct {
	numBytes  byte
	textBytes [maxBytesPerLeaf]byte
}

func (l *leafNode) key() indexKey {
	key := indexKey{}
	for _, b := range l.textBytes[:l.numBytes] {
		key.numChars += uint64(utf8StartByteIndicator[b])
		if b == '\n' {
			key.numLines++
		}
	}
	return key
}

func (l *leafNode) byteOffsetForPosition(charPos uint64) byte {
	n := uint64(0)
	for i, b := range l.textBytes[:l.numBytes] {
		c := utf8StartByteIndicator[b]
		if c > 0 && n == charPos {
			return byte(i) // safe b/c maxBytesPerLeaf < 256
		}
		n += uint64(c)
	}
	return l.numBytes
}

// Lookup table for UTF-8 bytes. Set to 1 for start bytes, zero otherwise.
var utf8StartByteIndicator [256]byte

func init() {
	for b := 0; b < 256; b++ {
		if b>>7 == 0 ||
			b>>5 == 0b110 ||
			b>>4 == 0b1110 ||
			b>>3 == 0b11110 {
			utf8StartByteIndicator[b] = 1
		}
	}
}

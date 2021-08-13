package rtree

import (
	"math"
	"math/bits"
)

const (
	minChildren = 2
	maxChildren = 4
)

func (t *RTree) Insert(box Box, recordID int) {
	t.count++

	if t.root == nil {
		t.root = &node{isLeaf: true}
	}

	level := t.root.depth() - 1
	leaf := t.chooseBestNode(box, level)

	leaf.appendRecord(box, recordID)
	t.adjustBoxesUpwards(leaf, box)

	if leaf.numEntries <= maxChildren {
		return
	}

	newNode := t.splitNode(leaf)
	root1, root2 := t.adjustTree(leaf, newNode)
	if root2 != nil {
		t.joinRoots(root1, root2)
	}
}

func (t *RTree) adjustBoxesUpwards(node *node, box Box) {
	for node != t.root {
		parent := node.parent
		for i := 0; i < parent.numEntries; i++ {
			e := &parent.entries[i]
			if e.child == node {
				e.box = combine(e.box, box)
			}
		}
		node = parent
	}
}

func (t *RTree) joinRoots(r1, r2 *node) {
	newRoot := &node{
		entries: [1 + maxChildren]entry{
			{box: calculateBound(r1), child: r1},
			{box: calculateBound(r2), child: r2},
		},
		numEntries: 2,
		parent:     nil,
		isLeaf:     false,
	}
	r1.parent = newRoot
	r2.parent = newRoot
	t.root = newRoot
}

func (t *RTree) adjustTree(n, nn *node) (*node, *node) {
	for {
		if n == t.root {
			return n, nn
		}
		parent := n.parent
		for i := 0; i < parent.numEntries; i++ {
			if parent.entries[i].child == n {
				parent.entries[i].box = calculateBound(n)
				break
			}
		}

		var pp *node
		if nn != nil {
			parent.appendChild(calculateBound(nn), nn)
			if parent.numEntries > maxChildren {
				pp = t.splitNode(parent)
			}
		}

		n, nn = parent, pp
	}
}

func (t *RTree) splitNode(n *node) *node {
	var (
		minSplit = uint64(1)
		maxSplit = uint64((1 << (n.numEntries - 1)) - 1)
	)
	bestArea := math.Inf(+1)
	var bestSplit uint64
	for split := minSplit; split <= maxSplit; split++ {
		if ones := bits.OnesCount64(split); ones < minChildren || (n.numEntries-ones) < minChildren {
			continue
		}
		var boxA, boxB Box
		var hasA, hasB bool
		for i := 0; i < n.numEntries; i++ {
			entryBox := n.entries[i].box
			if split&(1<<i) == 0 {
				if hasA {
					boxA = combine(boxA, entryBox)
				} else {
					boxA = entryBox
				}
			} else {
				if hasB {
					boxB = combine(boxB, entryBox)
				} else {
					boxB = entryBox
				}
			}
		}
		combinedArea := area(boxA) + area(boxB)
		if combinedArea < bestArea {
			bestArea = combinedArea
			bestSplit = split
		}
	}

	newNode := &node{isLeaf: n.isLeaf}
	totalEntries := n.numEntries
	n.numEntries = 0
	for i := 0; i < totalEntries; i++ {
		entry := n.entries[i]
		if bestSplit&(1<<i) == 0 {
			n.entries[n.numEntries] = entry
			n.numEntries++
		} else {
			newNode.entries[newNode.numEntries] = entry
			newNode.numEntries++
		}
	}
	for i := n.numEntries; i < len(n.entries); i++ {
		n.entries[i] = entry{}
	}
	if !n.isLeaf {
		for i := 0; i < newNode.numEntries; i++ {
			newNode.entries[i].child.parent = newNode
		}
	}
	return newNode
}

func (t *RTree) chooseBestNode(box Box, level int) *node {
	node := t.root
	for {
		if level == 0 {
			return node
		}
		bestDelta := enlargement(box, node.entries[0].box)
		bestEntry := 0
		for i := 1; i < node.numEntries; i++ {
			entryBox := node.entries[i].box
			delta := enlargement(box, entryBox)
			if delta < bestDelta {
				bestDelta = delta
				bestEntry = i
			} else if delta == bestDelta && area(entryBox) < area(node.entries[bestEntry].box) {
				bestEntry = i
			}
		}
		node = node.entries[bestEntry].child
		level--
	}
}

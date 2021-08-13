package rtree

import (
	"errors"
)

type node struct {
	entries    [1 + maxChildren]entry
	numEntries int
	parent     *node
	isLeaf     bool
}

type entry struct {
	box      Box
	child    *node
	recordID int
}

func (n *node) appendRecord(box Box, recordID int) {
	n.entries[n.numEntries] = entry{box: box, recordID: recordID}
	n.numEntries++
}

func (n *node) appendChild(box Box, child *node) {
	n.entries[n.numEntries] = entry{box: box, child: child}
	n.numEntries++
	child.parent = n
}

func (n *node) depth() int {
	var d = 1
	for !n.isLeaf {
		d++
		n = n.entries[0].child
	}
	return d
}

type RTree struct {
	root  *node
	count int
}

var Stop = errors.New("stop")

func (t *RTree) RangeSearch(box Box, callback func(recordID int) error) error {
	if t.root == nil {
		return nil
	}
	var recurse func(*node) error
	recurse = func(n *node) error {
		for i := 0; i < n.numEntries; i++ {
			entry := n.entries[i]
			if !overlap(entry.box, box) {
				continue
			}
			if n.isLeaf {
				if err := callback(entry.recordID); err == Stop {
					return nil
				} else if err != nil {
					return err
				}
			} else {
				if err := recurse(entry.child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return recurse(t.root)
}

func (t *RTree) Extent() (Box, bool) {
	if t.root == nil || t.root.numEntries == 0 {
		return Box{}, false
	}
	return calculateBound(t.root), true
}

func (t *RTree) Count() int {
	return t.count
}

package rtree

import "container/heap"

func (t *RTree) Nearest(box Box) (int, bool) {
	var (
		recordID int
		found    bool
	)
	t.PrioritySearch(box, func(rid int) error {
		recordID = rid
		found = true
		return Stop
	})
	return recordID, found
}

func (t *RTree) PrioritySearch(box Box, callback func(recordID int) error) error {
	if t.root == nil {
		return nil
	}

	queue := entriesQueue{origin: box}
	equeueNode := func(n *node) {
		for i := 0; i < n.numEntries; i++ {
			heap.Push(&queue, &n.entries[i])
		}
	}

	equeueNode(t.root)
	for len(queue.entries) > 0 {
		nearest := heap.Pop(&queue).(*entry)
		if nearest.child == nil {
			if err := callback(nearest.recordID); err != nil {
				if err == Stop {
					return nil
				}
				return err
			}
		} else {
			equeueNode(nearest.child)
		}
	}
	return nil
}

type entriesQueue struct {
	entries []*entry
	origin  Box
}

func (q *entriesQueue) Len() int {
	return len(q.entries)
}

func (q *entriesQueue) Less(i int, j int) bool {
	e1 := q.entries[i]
	e2 := q.entries[j]
	return squaredEuclideanDistance(e1.box, q.origin) < squaredEuclideanDistance(e2.box, q.origin)
}

func (q *entriesQueue) Swap(i int, j int) {
	q.entries[i], q.entries[j] = q.entries[j], q.entries[i]
}

func (q *entriesQueue) Push(x interface{}) {
	q.entries = append(q.entries, x.(*entry))
}

func (q *entriesQueue) Pop() interface{} {
	e := q.entries[len(q.entries)-1]
	q.entries = q.entries[:len(q.entries)-1]
	return e
}

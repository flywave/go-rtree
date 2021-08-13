package rtree

type BulkItem struct {
	Box      Box
	RecordID int
}

func BulkLoad(items []BulkItem) *RTree {
	if len(items) == 0 {
		return &RTree{}
	}

	levels := calculateLevels(len(items))
	return &RTree{bulkInsert(items, levels), len(items)}
}

func calculateLevels(numItems int) int {
	levels := 1
	count := maxChildren
	for count < numItems {
		count *= maxChildren
		levels++
	}
	return levels
}

func bulkInsert(items []BulkItem, levels int) *node {
	if levels == 1 {
		root := &node{isLeaf: true, numEntries: len(items)}
		for i, item := range items {
			root.entries[i] = entry{
				box:      item.Box,
				recordID: item.RecordID,
			}
		}
		return root
	}

	if len(items) < 6 {
		firstHalf, secondHalf := splitBulkItems2Ways(items)
		return bulkNode(levels, firstHalf, secondHalf)
	}

	if len(items) < 8 {
		firstThird, secondThird, thirdThird := splitBulkItems3Ways(items)
		return bulkNode(levels, firstThird, secondThird, thirdThird)
	}

	firstHalf, secondHalf := splitBulkItems2Ways(items)
	firstQuarter, secondQuarter := splitBulkItems2Ways(firstHalf)
	thirdQuarter, fourthQuarter := splitBulkItems2Ways(secondHalf)
	return bulkNode(levels, firstQuarter, secondQuarter, thirdQuarter, fourthQuarter)
}

func bulkNode(levels int, parts ...[]BulkItem) *node {
	root := &node{
		numEntries: len(parts),
		parent:     nil,
		isLeaf:     false,
	}
	for i, part := range parts {
		child := bulkInsert(part, levels-1)
		child.parent = root
		root.entries[i].child = child
		root.entries[i].box = calculateBound(child)
	}
	return root
}

func splitBulkItems2Ways(items []BulkItem) ([]BulkItem, []BulkItem) {
	horizontal := itemsAreHorizontal(items)
	split := len(items) / 2
	quickPartition(items, split, horizontal)
	return items[:split], items[split:]
}

func splitBulkItems3Ways(items []BulkItem) ([]BulkItem, []BulkItem, []BulkItem) {
	if ln := len(items); ln != 6 && ln != 7 {
		panic(len(items))
	}

	horizontal := itemsAreHorizontal(items)
	quickPartition(items, 2, horizontal)
	quickPartition(items[3:], 1, horizontal)

	return items[:2], items[2:4], items[4:]
}

func quickPartition(items []BulkItem, k int, horizontal bool) {
	var rndState uint32
	rnd := func(n int) int {
		rndState = 1664525*rndState + 1013904223
		return int((uint64(rndState) * uint64(n)) >> 32)
	}

	less := func(i, j int) bool {
		bi := items[i].Box
		bj := items[j].Box
		if horizontal {
			return bi.MinX+bi.MaxX < bj.MinX+bj.MaxX
		}
		return bi.MinY+bi.MaxY < bj.MinY+bj.MaxY
	}
	swap := func(i, j int) {
		items[i], items[j] = items[j], items[i]
	}

	left, right := 0, len(items)-1
	for {
		switch right - left {
		case 1:
			if less(right, left) {
				swap(right, left)
			}
			return
		case 2:
			if less(left+1, left) {
				swap(left+1, left)
			}
			if less(left+2, left+1) {
				swap(left+2, left+1)
				if less(left+1, left) {
					swap(left+1, left)
				}
			}
			return
		}

		pivot := left + rnd(right-left+1)
		if pivot != right {
			swap(pivot, right)
		}

		j := left
		for i := left; i < right; i++ {
			if less(i, right) {
				swap(i, j)
				j++
			}
		}

		swap(right, j)

		switch {
		case j-left < k:
			k -= j - left + 1
			left = j + 1
		case j-left > k:
			right = j - 1
		default:
			return
		}
	}
}

func itemsAreHorizontal(items []BulkItem) bool {
	minX := items[0].Box.MinX
	maxX := items[0].Box.MaxX
	minY := items[0].Box.MinY
	maxY := items[0].Box.MaxY
	for _, item := range items[1:] {
		box := item.Box
		minX = fastMin(minX, box.MinX)
		maxX = fastMax(maxX, box.MaxX)
		minY = fastMin(minY, box.MinY)
		maxY = fastMax(maxY, box.MaxY)
	}
	return maxX-minX > maxY-minY
}

func fastMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func fastMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

package rtree

type Box struct {
	MinX, MinY, MaxX, MaxY float64
}

func calculateBound(n *node) Box {
	box := n.entries[0].box
	for i := 1; i < n.numEntries; i++ {
		box = combine(box, n.entries[i].box)
	}
	return box
}

func combine(box1, box2 Box) Box {
	return Box{
		MinX: fastMin(box1.MinX, box2.MinX),
		MinY: fastMin(box1.MinY, box2.MinY),
		MaxX: fastMax(box1.MaxX, box2.MaxX),
		MaxY: fastMax(box1.MaxY, box2.MaxY),
	}
}

func enlargement(existing, additional Box) float64 {
	return area(combine(existing, additional)) - area(existing)
}

func area(box Box) float64 {
	return (box.MaxX - box.MinX) * (box.MaxY - box.MinY)
}

func overlap(box1, box2 Box) bool {
	return true &&
		(box1.MinX <= box2.MaxX) && (box1.MaxX >= box2.MinX) &&
		(box1.MinY <= box2.MaxY) && (box1.MaxY >= box2.MinY)
}

func squaredEuclideanDistance(b1, b2 Box) float64 {
	dx := fastMax(0, fastMax(b1.MinX-b2.MaxX, b2.MinX-b1.MaxX))
	dy := fastMax(0, fastMax(b1.MinY-b2.MaxY, b2.MinY-b1.MaxY))
	return dx*dx + dy*dy
}

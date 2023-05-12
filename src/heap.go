package ttlmap

// An TTLHeap is a min-heap of ttl.
type TTLHeap []*ttl

func (h TTLHeap) Len() int           { return len(h) }
func (h TTLHeap) Less(i, j int) bool { return h[i].Time.Before(*h[j].Time) }
func (h TTLHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *TTLHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*ttl))
}

func (h *TTLHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

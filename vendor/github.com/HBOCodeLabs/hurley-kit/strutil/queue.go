package strutil

// Queue is a FIFO queue data structure
type Queue []string

// Push adds an item to the end of the queue
func (q *Queue) Push(s string) {
	*q = append(*q, s)
}

// Pop removes an item from the head of the queue
func (q *Queue) Pop() string {
	rv := (*q)[0]
	*q = (*q)[1:]
	return rv
}

// Len returns the length of the queue
func (q *Queue) Len() int {
	return len(*q)
}

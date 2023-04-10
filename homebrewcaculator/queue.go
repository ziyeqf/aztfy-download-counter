package homebrewcaculator

type queue struct {
	Head *node
	Tail *node
	Size int
}

type node struct {
	Value calcFunc
	Next  *node
}

func (q *queue) enqueue(v ...calcFunc) {
	if len(v) == 0 {
		return
	}
	for _, v := range v {
		q.Size++
		n := &node{Value: v}
		if q.Head == nil {
			q.Head = n
			q.Tail = n
			continue
		}
		q.Tail.Next = n
		q.Tail = n
	}
}

func (q *queue) dequeue() calcFunc {
	if q.Head == nil {
		return nil
	}
	v := q.Head.Value
	q.Head = q.Head.Next
	q.Size--
	return v
}

func (q *queue) isEmpty() bool {
	return q.Head == nil
}

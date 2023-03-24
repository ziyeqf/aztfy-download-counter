package homebrewcaculator

type Queue struct {
	Head *Node
	Tail *Node
}

type Node struct {
	Value calcFunc
	Next  *Node
}

func (q *Queue) Enqueue(v ...calcFunc) {
	if len(v) == 0 {
		return
	}
	for _, v := range v {
		n := &Node{Value: v}
		if q.Head == nil {
			q.Head = n
			q.Tail = n
			return
		}
		q.Tail.Next = n
		q.Tail = n
	}
}

func (q *Queue) Dequeue() calcFunc {
	if q.Head == nil {
		return nil
	}
	v := q.Head.Value
	q.Head = q.Head.Next
	return v
}

func (q *Queue) IsEmpty() bool {
	return q.Head == nil
}

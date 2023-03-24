package homebrewcaculator

type Queue struct {
	Head *Node
	Tail *Node
}

type Node struct {
	Value CalculateDataAtDayNFunc
	Next  *Node
}

func (q *Queue) Enqueue(v ...CalculateDataAtDayNFunc) {
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

func (q *Queue) Dequeue() *CalculateDataAtDayNFunc {
	if q.Head == nil {
		return nil
	}
	v := q.Head.Value
	q.Head = q.Head.Next
	return &v
}

func (q *Queue) IsEmpty() bool {
	return q.Head == nil
}

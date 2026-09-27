package utils

type CircularQueue struct {
	data  []any
	size  int
	front int
	tail  int
}

func NewCircularQueue(capacity int) *CircularQueue {
	queue := &CircularQueue{
		data:  make([]any, capacity),
		size:  0,
		front: 0,
		tail:  0,
	}
	return queue
}

func (q *CircularQueue) Enqueue(item any) bool {
	if q.size == len(q.data) {
		return false // 队列已满
	}
	q.data[q.tail] = item
	q.tail = (q.tail + 1) % len(q.data)
	q.size++
	return true
}

func (q *CircularQueue) Dequeue() any {
	if q.size == 0 {
		return nil // 队列为空
	}
	item := q.data[q.front]
	q.front = (q.front + 1) % len(q.data)
	q.size--
	return item
}

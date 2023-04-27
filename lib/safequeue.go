package lib

import "sync"

type SafeQueue[T any] struct {
	queue []T
	mutex *sync.Mutex
	done  bool
}

func NewSafeQueue[T any]() *SafeQueue[T] {
	return &SafeQueue[T]{queue: make([]T, 0), mutex: &sync.Mutex{}, done: false}
}

func (q *SafeQueue[T]) Add(ts ...T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.queue = append(q.queue, ts...)
}

func (q *SafeQueue[T]) Pop() T {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	ts := q.queue[0]
	q.queue = q.queue[1:]
	return ts
}

func (q *SafeQueue[T]) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.queue)
}

func (q *SafeQueue[T]) IsEmpty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.queue) == 0
}

func (q *SafeQueue[T]) SetDone(canBeDone bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.done = canBeDone
}

func (q *SafeQueue[T]) IsDone() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.done
}

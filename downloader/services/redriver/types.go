package redriver

import "github.com/gcottom/semaphore"

type ReDriverService interface {
	Add(id string)
	DeQueue() string
}

type Service struct {
	DB    map[string]int
	Lock  *semaphore.Semaphore
	Queue []string
}

func NewService() *Service {
	lock := semaphore.NewSemaphore(2)
	lock.Acquire()
	return &Service{
		DB:    make(map[string]int),
		Lock:  lock,
		Queue: make([]string, 0),
	}
}

package redriver

func (s *Service) Add(id string) {
	s.Lock.Acquire()
	defer s.Lock.Release()
	if count, exist := s.DB[id]; exist {
		s.DB[id] = count + 1
	} else {
		s.DB[id] = 1
	}
	if s.DB[id] <= 5 {
		s.Queue = append(s.Queue, id)
	}
}

func (s *Service) DeQueue() string {
	s.Lock.Acquire()
	defer s.Lock.Release()
	if len(s.Queue) == 0 {
		return ""
	}
	id := s.Queue[0]
	s.Queue = s.Queue[1:]
	return id
}

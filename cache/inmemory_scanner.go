package cache

type inMemoryScanner struct {
	pair
	pairChan  chan *pair
	closeChan chan struct{}
}

func (s *inMemoryScanner) Close() {
	close(s.closeChan)
}

func (s *inMemoryScanner) Key() string {
	return s.pair.k
}

func (s *inMemoryScanner) Value() []byte {
	return s.pair.v
}

func (s *inMemoryScanner) Scan() bool {
	p, ok := <-s.pairChan
	if ok {
		s.pair.k, s.pair.v = p.k, p.v
	}

	return ok
}

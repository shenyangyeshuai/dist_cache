package cache

import (
	"sync"
	"time"
)

type value struct {
	v       []byte
	created time.Time
}

type inMemoryCache struct {
	c     map[string]value
	mutex sync.RWMutex
	Stat
	ttl time.Duration
}

func newInMemoryCache(ttl int) *inMemoryCache {
	ca := &inMemoryCache{
		c:     make(map[string]value),
		mutex: sync.RWMutex{},
		Stat:  Stat{},
		ttl:   time.Duration(ttl) * time.Second,
	}

	if ttl > 0 {
		go ca.expirer()
	}

	return ca
}

func (c *inMemoryCache) expirer() {
	for {
		time.Sleep(c.ttl)
		c.mutex.RLock()
		for k, v := range c.c {
			c.mutex.RUnlock()
			if v.created.Add(c.ttl).Before(time.Now()) {
				c.Del(k)
			}
			c.mutex.RLock()
		}
		c.mutex.RUnlock()
	}
}

func (c *inMemoryCache) Set(k string, v []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Stat.add(k, v)
	c.c[k] = value{v: v, created: time.Now()}
	return nil
}

func (c *inMemoryCache) Get(k string) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.c[k].v, nil
}

func (c *inMemoryCache) Del(k string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	v, exist := c.c[k]
	if exist {
		delete(c.c, k)
		c.del(k, v.v)
	}
	return nil
}

func (c *inMemoryCache) GetStat() Stat {
	return c.Stat
}

func (c *inMemoryCache) NewScanner() Scanner {
	pairChan := make(chan *pair)
	closeChan := make(chan struct{})

	go func() {
		defer close(pairChan)
		c.mutex.RLock()
		for k, v := range c.c {
			c.mutex.RUnlock()
			select {
			case <-closeChan:
				return
			case pairChan <- &pair{k, v.v}:
			}
			c.mutex.RLock()
		}

		c.mutex.RUnlock()
	}()

	return &inMemoryScanner{pair{}, pairChan, closeChan}
}

package api

import "sync"

//Cache is a cache id lookup by MD5 hash
type Cache struct {
	store map[Hash]int
	mu    *sync.RWMutex
}

//Add adds an entry to the cache
func (c *Cache) Add(h Hash, id int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[h] = id
}

//Delete removes an entry from the cache
func (c *Cache) Delete(h Hash) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, h)
}

//Get returns an id by hash. If the entry doesn't exist, 0 is returned
func (c *Cache) Get(h Hash) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	id := c.store[h]
	return id
}

//Visit calls fn(key, val) for each entry in the cache.
//fn should not attempt any modifications to key or val, and should not call any methods on c
func (c *Cache) Visit(fn func(key Hash, val int)) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for key, val := range c.store {
		fn(key, val)
	}
}

//Clear removes all entries from th  cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.store {
		delete(c.store, i)
	}
}

//NewCache returns a new initialized cache
func NewCache() *Cache {
	return &Cache{
		store: make(map[Hash]int),
		mu:    new(sync.RWMutex),
	}
}

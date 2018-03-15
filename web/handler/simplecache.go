package handler

import (
	"sync"
	"time"
)

type SimpleExpireableCacheItem interface {
	IsExpired() bool
}

type SimpleCacheItem struct {
	storeTime time.Time
	expires   int
	content   string
}

func (item *SimpleCacheItem) IsExpired() bool {
	if item.expires <= 0 {
		return false
	}
	return time.Now().Sub(item.storeTime) > time.Second*time.Duration(item.expires)
}

type SimpleCache struct {
	items sync.Map
}

func (c *SimpleCache) Set(key interface{}, value SimpleExpireableCacheItem) {
	c.items.Store(key, value)
}

func (c *SimpleCache) Get(key interface{}) (interface{}, bool) {
	v, ok := c.items.Load(key)
	if !ok {
		return nil, false
	}
	item, _ := v.(SimpleExpireableCacheItem)
	if item.IsExpired() {
		c.Delete(key)
		return nil, false
	}
	return item, true
}

func (c *SimpleCache) Delete(key interface{}) {
	c.items.Delete(key)
}

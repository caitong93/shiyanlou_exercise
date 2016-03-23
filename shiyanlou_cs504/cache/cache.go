package cache

import (
	"fmt"
	"sync"
	"time"
)

const (
	// 没有过期时间标志
	NoExpiration time.Duration = -1

	// 默认的过期时间
	DefaultExpiration time.Duration = 0
)

type Item struct {
	Object     interface{} // 真正的数据项
	Expiration int64       // 生存时间
}

// 判断数据项是否已经过期
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

type Cache struct {
	defaultExpiration time.Duration
	items             map[string]Item // 缓存数据项存储在 map 中
	mu                sync.RWMutex    // 读写锁
	gcInterval        time.Duration   // 过期数据项清理周期
	stopGc            chan bool
}

// 过期缓存数据项清理
func (c *Cache) gcLoop() {
	ticker := time.NewTicker(c.gcInterval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-c.stopGc:
			ticker.Stop()
			return
		}
	}
}

// 删除缓存数据项
func (c *Cache) delete(k string) {
	delete(c.items, k)
}

// 删除过期数据
func (c *Cache) DeleteExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			c.delete(k)
		}
	}
}

// 删除一个数据项
func (c *Cache) Delete(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delete(k)
}

// 设置缓存数据项, 如果存在则覆盖，没有锁操作
func (c *Cache) set(k string, v interface{}, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.items[k] = Item{
		Object:     v,
		Expiration: e,
	}
}

// 设置缓存数据项, 如果存在则覆盖
func (c *Cache) Set(k string, v interface{}, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(k, v, d)
}

// 获取数据项。如果找到，还要判断是否过期
func (c *Cache) Add(k string, v interface{}, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, found := c.items[k]
	if found {
		return fmt.Errorf("Item %s already exists.", k)
	}
	c.set(k, v, d)
	return nil
}

// 获取数据项
func (c *Cache) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return nil, false
	}
	if item.Expired() {
		return item, false
	}
	return item, true
}

// 创建一个缓存系统
func NewCache(defaultExpiration, gcInterval time.Duration) *Cache {
	c := &Cache{
		defaultExpiration: defaultExpiration,
		gcInterval:        gcInterval,
		items:             map[string]Item{},
	}
	// 开始启动过期清理 goroutine
	go c.gcLoop()
	return c
}

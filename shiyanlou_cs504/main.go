package main

import (
	"./cache"
	"fmt"
	"time"
)

func main() {
	defaultExpiration, _ := time.ParseDuration("0.5h")
	gcInterval, _ := time.ParseDuration("1s")

	c := cache.NewCache(defaultExpiration, gcInterval)

	item1 := "qwertyuiop"
	expiration, _ := time.ParseDuration("5s")

	c.Set("k1", item1, expiration)

	time.Sleep(2 * 1e9)
	v, ok := c.Get("k1")
	fmt.Println(v, ok)

	err := c.Add("k1", "dummy", 0)
	fmt.Println(err)

	c.Set("k1", "dummy dummy", 1*1e9)
	v, ok = c.Get("k1")
	fmt.Println(v, ok)

	time.Sleep(3 * 1e9)
	v, ok = c.Get("k1")
	fmt.Println(v, ok)
}

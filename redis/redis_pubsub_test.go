package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"testing"
	"time"
)

func TestSub(t *testing.T) {
	c := redis.NewClient(&redis.Options{DB: 0, Addr: "localhost:6379"})
	defer c.Close()

	go sub("1", c)

	c.SetNX("1", 1, time.Minute)

	time.Sleep(time.Second * 3)

	c.Del("1")
	c.Publish("1", "unlocked")

	time.Sleep(time.Second)
}

func sub(key string, c *redis.Client) {
	sub := c.Subscribe(key)
	iface, err := sub.Receive()
	if err != nil {
		panic(err)
	}
	defer sub.Close()

	fmt.Println(iface)

	for n := range sub.Channel() {
		fmt.Println(*n)
	}
}

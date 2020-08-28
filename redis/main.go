package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

import (
	"github.com/go-redis/redis"
)

// 多个 processor 在处理前需要获得锁，否则轮询获取
// - 在解锁前，先判断是否是自己拥有的锁
// - 未获得锁，轮询获取，或参见 redis_pubsub_test.go 通过发布订阅方式
// 未处理情况：
// 1. 拥有锁的 processor 在 timeout 到期之前，事情还没做完（严重）
// 1.1 将过期时间设置得足够长，确保代码可以执行完
// 1.2 为获取锁的协程增加守护进程，为将要过期但未释放的锁增加有效时间
// 2. 拥有锁的 processor 处理过程中，挂掉了，others 需要白白等待

func main() {
	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		j := i
		go process(strconv.Itoa(j), &wg)
	}
	wg.Wait()
}

func process(clientID string, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	lockerKey := "locker1"
	addr := "localhost:6379"
	c := redis.NewClient(&redis.Options{DB: 0, Addr: addr})
	fmt.Printf("%s getting %s=%s\n", clientID, lockerKey, clientID)

	for {
		ok, err := c.SetNX(lockerKey, clientID, time.Second*100).Result()
		if err != nil {
			fmt.Println(clientID, "failed to get ", lockerKey)
			return
		}
		if ok {
			break
		}
		time.Sleep(time.Second)
	}

	fmt.Println(clientID, "owning ", lockerKey)

	// process something...

	time.Sleep(time.Second * 3)

	// check if the locker is mine
	val := c.Get(lockerKey).Val()
	if val == clientID {
		fmt.Println(clientID, "delete", lockerKey)
		c.Del(lockerKey)
	} else {
		fmt.Printf("the value is %#v can't del by %s\n", val, clientID)
	}
}

package main

import (
	"github.com/go-redis/redis"
	"testing"
)

var client2 *redis.Client

func prepare() {
	// init redis
	client2 = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client2.Ping().Result()
	if err != nil {
		println("redis connected failed")
	} else {
		println("redis connected success")
	}
	// Output: PONG <nil>
}

func TestPFAdd(t *testing.T) {
	prepare()
	res := client2.PFAdd("testuv", "abcd")
	res2 := client2.PFAdd("testuv", "abcd")
	// 如果返回1，表示没有该元素
	t.Log(res.Result())
	// 如果返回0，表示有该元素
	t.Log(res2.Result())
}

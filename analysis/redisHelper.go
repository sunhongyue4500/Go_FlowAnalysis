package main

import (
	"github.com/go-redis/redis"
	"net/url"
	"strings"
)

var client *redis.Client

func init() {
	// init redis
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		println("redis connected failed")
	} else {
		println("redis connected success")
	}
	// Output: PONG <nil>
}

func storeDataBlock(block *storageBlock) {
	//log.Printf("\r\nblock: %v\r\n", block)
	//sb = storageBlock{"pv", "ZINCREIBY", pvData.node}
	node := block.unode
	//  http://localhost:8888/movie/981.html
	//println("nodeurl:", node.unUrl)
	temp, _ := url.Parse(node.unUrl)
	splitPathArray := strings.Split(temp.Path[1:], "/")
	//println(temp.Path, t[0], t[1])
	// 遍历[movie, 4305.html]
	str := splitPathArray[0]
	for i:=1; i<len(splitPathArray); i++ {
		// 对每一层洋葱皮做统计
		str = str + "_" + splitPathArray[i]
		client.ZIncrBy(block.counterType, 1.0, str)
		client.ZIncrBy(block.counterType, 1.0, splitPathArray[i])
	}
}

package main

import (
	"flag"
	"fmt"
	"goLandTest/liuliangAnalysis/ua"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type resource struct {
	url    string
	target string
	start  int
	end    int
}

func ruleResource() []resource {
	var res []resource
	r1 := resource{
		url:    "http://localhost:8888/",
		target: "",
		start:  0,
		end:    0,
	}
	r2 := resource{
		url:    "http://localhost:8888/list/{$id}.html",
		target: "{$id}",
		start:  1,
		end:    21,
	}
	r3 := resource{
		url:    "http://localhost:8888/movie/{$id}.html",
		target: "{$id}",
		start:  1,
		end:    12924,
	}
	res = append(append(append(res, r1), r2), r3)
	return res
}

func buildURL(res []resource) []string {
	list := []string{}
	for _, resItem := range res {
		if len(resItem.target) == 0 {
			list = append(list, resItem.url)
		} else {
			// 需要做替换
			for i := resItem.start; i <= resItem.end; i++ {
				newUrl := strings.Replace(resItem.url, resItem.target, strconv.Itoa(i), -1)
				list = append(list, newUrl)
			}
		}
	}
	return list
}

func makeLog(current, refer, ua string) string {
	url := url.Values{}
	url.Set("time", "1")
	url.Set("url", current)
	url.Set("refer", refer)
	url.Set("ua", ua)
	paramStr := url.Encode()
	template := ` 127.0.0.1 - - [30/Jul/2019:15:17:40 +0800] "GET /dig?{$paramStr} HTTP/1.0" 304 0 "-" "{$ua}" "-" "123.57.237.212"`
	log := strings.Replace(template, "{$paramStr}", paramStr, -1)
	log = strings.Replace(log, "{$ua}", ua, -1)
	return log
}

// 产生随机数范围 [min, max)
func randInt(min, max int) int {
	if max < min {
		return max
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 随机数范围 [min, max)
	return r.Intn(max-min) + min
}

func main() {
	// 批量生产日志，通过程序代码生成任意多的日志
	defaultPath := "/Users/hongyi/Code/Practice/Go/src/goLandTest/liuliangAnalysis/dig.log"
	path := flag.String("path", defaultPath, "log path")
	row := flag.Int("row", 50, "log row nums")
	flag.Parse()
	fmt.Println(*path, *row)

	// 构造出真实网站的url集合
	res := ruleResource()
	list := buildURL(res)
	fmt.Println(list)

	uaList := ua.GetUA()

	fd, err := os.OpenFile(*path, os.O_CREATE | os.O_RDWR | os.O_APPEND, 0644)
	defer fd.Close()
	if err != nil {
		log.Printf("open file failed: %v\n", err)
	}

	for i := 0; i < *row; i++ {
		// 获取Fake数据
		currentUrl := list[randInt(0, len(list))]
		refer := list[randInt(0, len(list))]
		uaFake := uaList[randInt(0, len(uaList))]
		aLog := makeLog(currentUrl, refer, uaFake)
		fd.WriteString(aLog + "\n")
	}

	// 生成$row行日志到日志文件
	fmt.Println("done")
}

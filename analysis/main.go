package main

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const DIGSTART = " /dig?"
const HANDLE_MOIVE = "movie/"
const HANDLE_LIST = "list/"
const HANDLE_FIRST = "http://localhost:8888/"
const HANDLE_END = ".html"

// 读取的日志行数
var logLinesCounter = 0

// 包装命令行参数
type cmdParams struct {
	logPath string
	numGoroutine int
}

type digData struct {
	time string
	url string
	refer string
	ua string
}

type urlNode struct {
	unType string	// 类型，movie list
	unRid int	// 资源id
	unUrl string	// 当前这个页面的url
	unTime string	// 当前访问这个页面的时间
}

type urlData struct {
	data digData
	uid string
	node urlNode
}

type storageBlock struct {
	counterType string
	storageModel string
	unode urlNode
}

var loger logrus.Logger
func init()  {
	loger = *logrus.New()
	loger.Out = os.Stdout	// 默认是标准输出
	loger.Level = logrus.DebugLevel // 日志等级
}

func main() {
	// 该项目的日志文件
	defaultTargetLogPath := "/Users/hongyi/Code/Practice/Go/src/goLandTest/liuliangAnalysis/analysis/analysis.log"
	// 输入日志文件
	logFilePath := "/Users/hongyi/Code/Practice/Go/src/goLandTest/liuliangAnalysis/dig.log"
	// 并发的goroutine
	logAnalysiserGroutineNum := 5
	logFilePath = *flag.String("logFilePath", logFilePath, "log file path to analysis")
	logAnalysiserGroutineNum = *flag.Int("goroutineNum", logAnalysiserGroutineNum, "log analysis goroutine num")
	targetLogPath := *flag.String("targetLogPath", defaultTargetLogPath, "log file path to analysis")
	flag.Parse()
	//fmt.Println(logFilePath, logAnalysiserGroutineNum)

	params := cmdParams{logFilePath, logAnalysiserGroutineNum}
	// 打日志
	fd, err := os.OpenFile(targetLogPath, os.O_CREATE | os.O_RDWR | os.O_APPEND, 0644)
	if err == nil {
		loger.Out = fd
	} else {
		loger.Printf("open target log fail: %v\n", err)
	}
	//loger.Printf("logFile:%v, goroutineNum:%v", logFilePath, logAnalysiserGroutineNum)
	// 初始化channel用于传递数据
	bufNum := 10
	logAnalysisChannel := make(chan string, bufNum * 5)
	pvAnalysisChannel := make(chan urlData, bufNum)
	uvAnalysisChannel := make(chan urlData, bufNum)
	storeChannel := make(chan *storageBlock, bufNum)
	// 逐行读取日志文件
	go readFileLineByLine(params, logAnalysisChannel)
	// 创建一组日志处理
	for i:=0; i<logAnalysiserGroutineNum; i++ {
		go logAnalysis(logAnalysisChannel, pvAnalysisChannel, uvAnalysisChannel)
	}
	// 将读取的内容交给消费goroutines
	// pv， uv统计
	go pvCounter(pvAnalysisChannel, storeChannel)
	go uvCounter(uvAnalysisChannel, storeChannel)
	// 存储
	go store(storeChannel)

	time.Sleep(10 * time.Second)
}

// 读取日志数据
func readFileLineByLine(params cmdParams, logChanenl chan string) error {
	// 打开文件
	fd, err := os.OpenFile(params.logPath, os.O_CREATE | os.O_RDWR | os.O_APPEND, 0644)
	if err != nil {
		logrus.Warnf("open log file faied: %v", err)
	}
	defer fd.Close()	// 不要忘记关闭
	// 读取日志数据
	reader := bufio.NewReader(fd)
	for {
		line, err := reader.ReadString('\n')
		logLinesCounter++
		if err != nil {
			if err == io.EOF {
				// 读到末尾
				//time.Sleep(time.Second * 3)
				log.Printf("读取%d行日志\r\n", logLinesCounter)
				return nil
			} else {
				// 遇到错误
				log.Printf("读取%d行日志\r\n", logLinesCounter)
				loger.Errorf("Read log content fail: %v", err)
				return err
			}
		}
		//fmt.Print("读到数据：", line)	// 结尾带换行
		// 过滤无效数据
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		logChanenl <- line	// 放到channel
	}
	return nil
}

// 日志分析
func logAnalysis(log chan string, uv chan urlData, pv chan urlData) {
	for str := range log {
		// 切割str
		digDataTemp, _ := cutLog(str)
		//fmt.Println(urlDataTemp)
		// 模拟线上uid
		uid := sha256.Sum256([]byte(digDataTemp.refer + digDataTemp.ua))
		formatData := formatNode(digDataTemp.url, digDataTemp.time)
		urlDataTemp := urlData{digDataTemp, fmt.Sprintf("%x", uid), *formatData}
		uv <- urlDataTemp
		pv <- urlDataTemp
	}
}

// pv统计，来一个计一个(需要实现剥洋葱式统计)
func pvCounter(pv chan urlData, storeChan chan *storageBlock) {
	var sb storageBlock
	for pvData := range pv {
		//fmt.Println("收到pv", pvData)
		sb = storageBlock{"pv", "ZINCREIBY", pvData.node}
		storeChan <- &sb
	}
}

func uvCounter(uv chan urlData, storeChan chan *storageBlock) {
	// uv统计，需要去重，使用hyperLogLog去重
	var sb storageBlock
	for uvData := range uv {
		//fmt.Println("收到uv", uvData)
		// 需要按天去重
		val, _ := client.PFAdd("uv_hpll", uvData.data.url).Result()
		if val == 1{
			// 表示没有重复
			sb = storageBlock{"uv", "ZINCREIBY", uvData.node}
			storeChan <- &sb
		}
	}
}

func store(storeChan chan *storageBlock) {
	for dataBlock := range storeChan {
		storeDataBlock(dataBlock)
	}
}

func cutLog(str string) (digData, error) {
	startPos := strings.Index(str, DIGSTART)
	if startPos < 0 {
		// 没找到DIGSTART
		return digData{}, errors.New("cut log failed")
	}
	endPos := strings.Index(str, "HTTP/")
	if endPos < 0 {
		// 没找到DIGSTART
		return digData{}, errors.New("cut log failed")
	}
	//fmt.Println("cutK:", str[startPos:endPos])
	url, _ := url.Parse(str[startPos:endPos])
	va := url.Query()
	timeTemp := va["time"][0]
	urlTemp := va["url"][0]
	referTemp := va["refer"][0]
	uaTemp := va["ua"][0]
	return digData{timeTemp, urlTemp, referTemp, uaTemp}, nil
}

// 构造urlNode
func formatNode(url, t string) *urlNode {
	// movie > list >= 首页
	// 扣数据
	// 首先找movie
	pos1 := strings.Index(url, HANDLE_MOIVE)
	if pos1 != -1 {
		// 找到
		pos1 = pos1 + len(HANDLE_MOIVE)
		pos2 := strings.Index(url, HANDLE_END)
		if pos2 != -1 {
			id := url[pos1:pos2]
			_id, _ := strconv.Atoi(id)
			return &urlNode{"movie", _id, url, t}
		}
	}

	// 如果上述没找到，找list
	pos1 = strings.Index(url, HANDLE_LIST)
	if pos1 != -1 {
		// 找到
		pos1 = pos1 + len(HANDLE_LIST)
		pos2 := strings.Index(url, HANDLE_END)
		if pos2 != -1 {
			id := url[pos1+1:pos2]
			_id, _ := strconv.Atoi(id)
			return &urlNode{"list", _id, url, t}
		}
	}

	// 上述都没找到
	if url == HANDLE_FIRST {
		// 首页
		return &urlNode{"first", 1, url, t}
	}
	println("返回nil")
	return nil
}
package utils

import (
	"log"
	"runtime"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type ComputerInfo struct {
	Username       string
	Ipv4           string
	JavaServerHost string
	NumCPU         int
	ExeDir         string
	ServerPort     int
}

var computerInfo ComputerInfo

func GetComputerInfo() ComputerInfo {
	if computerInfo == (ComputerInfo{}) {
		log.Printf("重新加载配置文件信息！")
		numCPU := runtime.NumCPU()
		runtime.GOMAXPROCS(numCPU * 2)
		ipv4 := "127.0.0.1"
		exeDir := core.ExecPath()
		// 获取属性值
		username := "d41d8cd98f00b204e9800998ecf8427e"
		serverPort := 80
		javaServerHost := "192.168.31.236:8080"
		computerInfo = ComputerInfo{Username: username, Ipv4: ipv4, NumCPU: numCPU, JavaServerHost: javaServerHost,
			ExeDir: exeDir, ServerPort: serverPort}
		log.Printf("本机相关信息：%v", core.ToJsonString(computerInfo))
	}
	return computerInfo
}

func ClearCache() {
	computerInfo = ComputerInfo{}
}

func Summary() {
	ms := new(runtime.MemStats)
	runtime.ReadMemStats(ms)
	log.Printf("从系统获取到的总内存%dm\n", ms.Sys/MB)
	log.Printf("从系统获取到的总内存%dm\n", ms.Sys/MB)
	log.Printf("堆空闲(申请未使用)%dm\n", ms.HeapIdle/MB)
	log.Printf("使用中堆内存%dm\n", ms.HeapInuse/MB)
	log.Printf("系统分配作为运行栈的内存%dm\n", ms.StackSys/MB)
	log.Printf("正在使用栈内存%dm\n", ms.StackInuse/MB)
	log.Printf("活动的对象%d\n", ms.Mallocs-ms.Frees)
	log.Printf("GC调用次数%d\n", ms.NumGC)
	log.Printf("当前goroutines数量%d\n", runtime.NumGoroutine())
}

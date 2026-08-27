package utils

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"syscall"
	"time"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// OverrideDisk 文件大小-2G*size
// 文件个数 fileCount
func OverrideDisk(size int, fileCount int, disk string, slice *[]MEMUsage, speedSlice *[]Speed, du time.Duration) {
	*slice = append(*slice, MEMUsage{0, time.Now().UTC().Add(-du).String()})
	content := []byte(core.GetRandomString(1024 * 1024 * 2))
	globalStartTime := time.Now()
	randomFilePrefix := core.GetRandomString(10)
	lastCost := int64(0)
	//2G的n倍大小的文件
	for i := 0; i < fileCount; i++ {
		filepath := disk + ":\\" + randomFilePrefix + "." + strconv.Itoa(i)
		writeLargeFile(filepath, content, size)
		cost := time.Since(globalStartTime)
		costSecond := cost.Nanoseconds() - lastCost
		fmt.Printf("本次耗时：%v,_%v", costSecond, int64((2*size*(i+1))*1000*1000))
		speed := int64((2*size*(i+1))*1000*1000*1000) / costSecond
		lastCost = costSecond
		*slice = append(*slice, MEMUsage{2 * size * (i + 1), time.Now().UTC().Add(-du).String()})
		*speedSlice = append(*speedSlice, Speed{int(speed), time.Now().UTC().Add(-du).String()})
		fmt.Printf("-----------------%v", time.Now().Add(-du).UTC().String())
		fmt.Printf("已写入文件大小：%v,耗时：%v\n", strconv.Itoa(2*size*(i+1))+"M", cost.Seconds())
	}
	globalTimeCost := time.Since(globalStartTime)
	fmt.Printf("覆盖数据执行结束,耗时：%v\n", globalTimeCost)
}
func OverrideDisk1(size int, fileCount int, disk string, slice *[]MEMUsage) {
	for {
		time.Sleep(time.Duration(1) * time.Second)
		*slice = append(*slice, MEMUsage{4, time.Now().UTC().String()})
	}
}

// 写入2G*n大文件5
func writeLargeFile(filepath string, content []byte, n int) {
	f, err := os.OpenFile(filepath, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, os.ModeAppend)
	if err != nil {
		fmt.Printf("打开文件错误:%v", err.Error())
	}
	for i := 0; i < n; i++ {
		content[rand.Intn(len(content)-10)] = 'a'
		_, wrErr := f.Write(content)
		if wrErr != nil {
			fmt.Printf("写入错误:%v", wrErr.Error())
		}
	}
	if f.Close() != nil {
		fmt.Printf("文件关闭异常:%v", filepath)
	}
}

type Grafana struct {
	Data GrafanaData `json:"data"`
}

type GrafanaData struct {
	CPUUsage []CPUUsage `json:"CPU_Usage"`
	MEMUsage []MEMUsage `json:"MEM_Usage"`
	Speed    []Speed    `json:"Speed"`
}

type CPUUsage struct {
	CPUUsage int    `json:"CPU_Usage"`
	Time     string `json:"time"`
}

type MEMUsage struct {
	MEMUsage int    `json:"MEM_Usage"`
	Time     string `json:"time"`
}

type Speed struct {
	Speed int    `json:"Speed"`
	Time  string `json:"time"`
}

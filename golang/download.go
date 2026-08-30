package golang

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// DownloadFileWithUI 下载视频文件在目录中:workDir,map key-文件名 value-下载url
func DownloadFileWithUI(workDir string, myMap map[string]string, callback func([]byte)) {
	var wg sync.WaitGroup
	total := len(myMap)
	if total == 0 {
		callback([]byte("finish"))
		return
	}
	var done atomic.Int64
	for key, value := range myMap {
		wg.Add(1) // 增加WaitGroup的计数
		go func(key, value string) {
			defer wg.Done() // 当goroutine完成时，减少WaitGroup的计数
			DownloadM3u8(value, key, workDir, 48)
			n := done.Add(1)
			fmt.Printf("D_%d\n", (n*100)/int64(total))
			callback([]byte(fmt.Sprintf("D_%d", (n*100)/int64(total))))
		}(key, value)
	}
	// 等待所有goroutines完成
	wg.Wait()
	callback([]byte("finish"))
}

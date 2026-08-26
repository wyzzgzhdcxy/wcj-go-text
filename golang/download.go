package golang

import (
	"fmt"
	"sync"
)

// DownloadFileWithUI 下载视频文件在目录中:workDir,map key-文件名 value-下载url
func DownloadFileWithUI(workDir string, myMap map[string]string, callback func([]byte)) {
	var wg sync.WaitGroup
	total := len(myMap)
	i := 0
	for key, value := range myMap {
		wg.Add(1) // 增加WaitGroup的计数
		go func() {
			defer wg.Done() // 当goroutine完成时，减少WaitGroup的计数
			DownloadM3u8(value, key, workDir, 48)
			i++
			fmt.Printf("D_%d\n", (i*100)/total)
			callback([]byte(fmt.Sprintf("D_%d", (i*100)/total)))
		}()
	}
	// 等待所有goroutines完成
	wg.Wait()
	callback([]byte("finish"))
}

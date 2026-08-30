package cmdWrapper

import (
	"fmt"
	"strconv"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"

	"github.com/go-basic/uuid"
)

func DownloadM3u8(url string, fn string, dir string, threadCount int) {
	cacheTmp := core.GetTempDir() + "/ffmpeg/" + uuid.New()
	if core.FileExist(dir + "/" + fn) {
		fmt.Println("文件已经存在", dir+"/"+fn)
	} else {
		core.ExecuteCommand("m3u8", url, "--tmp-dir", cacheTmp, "--thread-count", strconv.Itoa(threadCount), "--no-log", "--save-name", fn, "--save-dir", dir)
	}
}
